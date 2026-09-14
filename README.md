# Dispatch City - Leistungsbeurteilung VSC-01

## Voraussetzungen

Die Lösung ist auf macOS und Windows mit Docker Desktop reproduzierbar.

- Docker Desktop mit aktivierter Kubernetes-Integration nicht nötig, es genügt
  eine laufende Docker Engine
- `k3d` >= 5.6
- `kubectl` >= 1.30
- `helm` >= 3.14
- `go` >= 1.22 (Build der Go-Services)
- `node` >= 20 mit `npm` (Build des Nuxt-Dashboards)
- `jq` für den Smoke-Test
- Internetzugang für Basis-Images sowie die Helm-Charts von RabbitMQ,
  CloudNativePG und `kube-prometheus-stack`

Standardnamen für den lokalen Cluster:

- Cluster: `delivery-lab`
- Kontext: `k3d-delivery-lab`
- Namespace: `food-delivery`
- Monitoring-Namespace: `monitoring`

## Repository-Struktur

```
apps/dashboard              Nuxt / PixiJS Dashboard (Block 3+)
cmd/                        Go-Entry-Points (control-api, worker, migrate, observer)
internal/                   Fachlogik, Messaging, Persistenz, Cluster-Observer
build/                      Gemeinsames Dockerfile für alle Go-Services
deploy/base                 Kustomize-Basis (Namespace, Deployments, Services, ConfigMap)
deploy/overlays/block-03-..-07  Kumulative Kustomize-Overlays pro Ausbaustufe
platform/cloudnative-pg     Helm-Installation des CloudNativePG-Operators
platform/monitoring         Helm-Installation von kube-prometheus-stack + Grafana-Setup
scripts/                    Build, Load, Smoke-Test, Reset, Skalierung, Failure-Demo
labs/block-07               Betriebs-Lab (Readiness, Rollout, HPA) inkl. Anleitung
docs/                       Architektur, Event-Katalog, Skalierungsdemo
```

## Ausbaustufen und Kustomize-Overlays

Die Overlays sind kumulativ; jedes höhere Overlay referenziert das
vorhergehende und erweitert es.

| Block | Fokus | Overlay | Sichtbares Ergebnis |
|---|---|---|---|
| 3 | Kubernetes Foundation | `deploy/overlays/block-03-standalone` | Dashboard und Control API als eigene Images, Port-Forward, In-Memory-Simulation |
| 4 | Ingress und Load Balancing | `deploy/overlays/block-04-ingress` | Traefik-Ingress unter `http://localhost:8080`, zwei Dashboard-Replikas, API singleton |
| 5 | RabbitMQ Event Pipeline | `deploy/overlays/block-05-messaging` | Producer / Consumer, Topic-Exchange, DLQ, `order.created` bis `order.delivered` |
| 6 | CloudNativePG Persistenz | `deploy/overlays/block-06-persistence` | Primary / Standby, Migration-Job, idempotente Order-Projektion |
| 7 | Observability und Resilienz | `deploy/overlays/block-07-observability` | Cluster-Observer, Prometheus / Grafana, HPA, Failure- und Skalierungsdemo |

## Build, Deploy und Verifikation

Alle Befehle laufen im Projektstamm. Die Befehle stellen den finalen Stand
(Block 7) her; frühere Ausbaustufen können durch den passenden Overlay-Namen
reproduziert werden.

### 1. Cluster und Operatoren bereitstellen

```bash
k3d cluster create delivery-lab \
  --api-port 6550 \
  -p "8080:80@loadbalancer" \
  --agents 1
sh platform/cloudnative-pg/install.sh
sh platform/monitoring/install.sh
```

### 2. Images bauen und in den Cluster laden

```bash
sh scripts/build-images.sh
sh scripts/load-images.sh
```

`build-images.sh` baut die Go-Services (Control API, Customer-, Restaurant-,
Courier-, Order-Worker, Migrate) über das gemeinsame Dockerfile in
`build/go-service.Dockerfile`. Das Dashboard-Image wird separat aus
`apps/dashboard` erwartet und beim ersten Start durch den mitgelieferten
Kursstand bereitgestellt. `load-images.sh` importiert alle Tags in den
k3d-Cluster.

### 3. Ziel-Overlay anwenden

Für den finalen Stand mit CloudNativePG und Observability:

```bash
sh scripts/deploy-final.sh
```

Für einen einzelnen Zwischenstand direkt mit Kustomize:

```bash
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-03-standalone
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-04-ingress
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-05-messaging
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-06-persistence
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-07-observability
```

### 4. Zugriffe

- Dashboard und API über Ingress: `http://localhost:8080`
- RabbitMQ Management (nach Block 5):

  ```bash
  kubectl --context k3d-delivery-lab -n food-delivery port-forward svc/rabbitmq 15672:15672
  ```

- Grafana (nach Block 7):

  ```bash
  kubectl --context k3d-delivery-lab -n monitoring port-forward svc/monitoring-grafana 3000:80
  ```

  Anmeldung: `admin` / `delivery`. Dashboard "Dispatch City - Betrieb".

### 5. Smoke-Test

```bash
sh scripts/smoke-test.sh
```

Der Smoke-Test prüft den Readiness-Endpunkt, den Snapshot (`mode ==
distributed`, drei Restaurants), einen manuellen Order-Trigger und wartet auf
Availability der Deployments, StatefulSets sowie den CloudNativePG-Cluster.

### 6. Reset

```bash
sh scripts/reset-demo.sh
```

Setzt den Simulationszustand über die Control-API zurück. Für ein vollständiges
Zurücksetzen der Persistenz muss zusätzlich der Overlay neu appliziert
werden.

### 7. Failure- und Skalierungsdemo

```bash
sh scripts/failure-demo.sh
sh scripts/scale-city.sh couriers 6
sh scripts/scale-city.sh customers 5
sh scripts/scale-city.sh restaurants 2
```

Details zur Beobachtung stehen in [docs/scaling-demo.md](docs/scaling-demo.md).
Das Betriebs-Lab aus Block 7 (Readiness, Rollout, HPA) inkl. Lasttest ist in
[labs/block-07/README.md](labs/block-07/README.md) beschrieben.

## Architektur (Kurzfassung)

Dashboard, Control API, RabbitMQ, die drei Simulator- bzw. Worker-Typen und der
Order-Worker sind eigenständige Deployments bzw. StatefulSets. Das Dashboard
konsumiert einen SSE-Stream der Control API sowie einen REST-Snapshot. Ab
Block 5 laufen fachliche Vorgänge als Events durch RabbitMQ; ab Block 6
projiziert der Order-Worker den Zustand in PostgreSQL. Details siehe
[docs/architecture.md](docs/architecture.md).

### Eventfluss

Zentrale Events: `customer.registered`, `courier.registered`, `order.created`,
`order.accepted`, `order.rejected`, `courier.assigned`,
`courier.location.updated`, `order.picked_up`, `order.delivered`,
`simulation.started`, `simulation.paused`, `simulation.reset`. Der
vollständige Katalog inklusive Producer, Consumer und Zweck steht in
[docs/event-catalog.md](docs/event-catalog.md). Jedes Event trägt `event_id`,
`event_version`, Zeitstempel, `correlation_id`, `causation_id` und `source`.
RabbitMQ liefert at-least-once; die Idempotenz stellt der Order-Worker über
die Tabelle `processed_events` sicher.

### Datenhaltung

Der Order-Worker ist der einzige Schreiber des fachlichen Order-Zustands. Eine
Verarbeitung besteht aus einer PostgreSQL-Transaktion (`event_id` in
`processed_events` beanspruchen, Projektion aktualisieren, relevantes Event in
`order_events` ablegen). Erst nach dem Commit wird die RabbitMQ-Nachricht
bestätigt. CloudNativePG stellt den `food-delivery-db-rw`-Service bereit, der
nach einem Failover automatisch auf den neuen Primary zeigt.

## Nachweis der finalen Demo

Screenshots und Kurzvideo der finalen Ausbaustufe liegen unter
[docs/demo/](docs/demo/) und decken die vom Bewertungsraster geforderten
Ansichten ab:

- Dashboard mit laufender Simulation, Systemansicht und Live-Events
- RabbitMQ Management UI (Exchanges, Queues, Backlog, DLQ)
- CloudNativePG-Cluster (`kubectl get cluster,pods -n food-delivery`) inkl.
  Failover-Ausgabe
- Grafana-Dashboard "Dispatch City - Betrieb" mit den relevanten Metriken
- Ausgabe von `scripts/smoke-test.sh` und `scripts/failure-demo.sh`

## Reflexion

- **Bekannte Grenzen**: Die Control API bleibt bewusst singleton, die horizontale Skalierung erfolgt nur über die Worker. Das
  Dashboard-Image wird aus dem Kursstand übernommen und nicht bei jedem
  `build-images.sh` neu gebaut.
- **Beobachtete Fehlerbilder**: Ohne `wait --for=condition=Ready` auf den
  CloudNativePG-Cluster kann der Migration-Job zu früh starten und läuft dann
  ins Timeout. Nach einem Neustart der RabbitMQ-Instanz verarbeitet der
  Order-Worker Backlog-Events erneut und die Idempotenz über `processed_events`
  verhindert doppelte Projektionen. Beim Herunterskalieren der
  Courier-Simulatoren stellt RabbitMQ nicht bestätigte Nachrichten erneut
  zu, offene Bestellungen bleiben bis zur Zustellung als `draining` sichtbar.
- **Mögliche Verbesserungen**: Aufteilung der Control API in einen zustands-
  losen HTTP-Adapter und einen zustandsbehafteten Steuer-Worker, damit auch
  die API horizontal skaliert werden kann. Vollständige CI mit Image-Build und
  automatischem Smoke-Test gegen einen k3d-Cluster.
  Alerting-Regeln in Grafana / Prometheus statt reiner Dashboards.

## Eingesetzte Hilfsmittel


- **Claude Code (Anthropic)**: Unterstützung beim Review des Repo, beim Formulieren dieses README aus den Arbeitsblättern und dem Repository-Zustand.
- **Von der Lehrperson bereitgestellte Kursstände sind unverändert übernommen und über die Commit-Historie
nachvollziehbar.**
