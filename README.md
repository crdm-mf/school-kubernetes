# Block 7: Kubernetes im Betrieb

Material zu Arbeitsblatt 07. Dieser Baustein erweitert den funktionierenden
Projektstand nach AB6. Er ist kein eigenstaendiges Projekt.

## Kursversion

[Release v1.1.0](https://github.com/SwitzerChees/vsc-dispatch-city-07-observability/releases/tag/v1.1.0)
passt zum aktuellen Arbeitsblatt 07.
[Kurs-ZIP herunterladen](https://github.com/SwitzerChees/vsc-dispatch-city-07-observability/releases/download/v1.1.0/vsc-dispatch-city-07-observability-v1.1.0.zip).

Das angehaengte Kurs-ZIP verwenden, nicht GitHubs automatisch erzeugtes
"Source code (zip)". Der enthaltene Ordner heisst
`vsc-dispatch-city-07-observability`; darin liegen `install.sh` und `install.ps1`.

## Inhalt

- Prometheus und Grafana als `kube-prometheus-stack`, Version `88.1.3`.
- Dashboard "Dispatch City - Betrieb" mit fuenf Anzeigen.
- Messpunkte fuer Anwendung, RabbitMQ und PostgreSQL.
- Cluster-Observer mit lesendem, auf `food-delivery` begrenztem RBAC.
- Kleines NGINX-Lab fuer Readiness, Rollback und HPA in `betrieb-lab`.

Der Einstieg benoetigt Docker Desktop, k3d, kubectl, Helm und Internet.
AB6 mit CloudNativePG muss bereits laufen. Der Monitoring-Stack ist fuer das
lokale Kurs-Lab reduziert; feste Zugangsdaten sind nicht fuer Produktion gedacht.

## In das bestehende Projekt integrieren

Entpackten Materialordner neben den Projektordner legen. Alle folgenden Befehle
im bestehenden Projektordner ausfuehren, jeweils als eigene Zeile.

Windows PowerShell:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
& '..\vsc-dispatch-city-07-observability\install.ps1' -Target '.'
./platform/monitoring/start-course.ps1
```

macOS oder WSL mit Bash:

```bash
sh ../vsc-dispatch-city-07-observability/install.sh .
sh platform/monitoring/start-course.sh
```

Die Integration kopiert nur die Dateien dieses Bausteins. Sie ersetzt weder
Dashboard noch vorhandene App-Dienste. Der Start baut und importiert das
Observer-Image, installiert den Helm-Stack und wendet das Block-7-Overlay an.
Standard: Cluster `teko-k8s`, Kontext `k3d-teko-k8s`.

Grafana in einem eigenen Terminal oeffnen:

```text
kubectl --context k3d-teko-k8s -n monitoring port-forward service/monitoring-grafana 3000:80
```

Browser: http://localhost:3000. Anmeldung: `admin` / `delivery`.
Unter Dashboards "Dispatch City - Betrieb" waehlen. Erste Daten brauchen
etwa eine Minute. "No data" ist kein Messwert von null.

## Die fuenf Aufgaben

1. Monitoring starten und zwei Anzeigen erklaeren.
2. Pizza-Worker auf null skalieren, Rueckstau beobachten und auf eins zurueckstellen.
3. `labs/block-07/web.yaml` anwenden, Readiness-Datei verschieben und wiederherstellen.
4. Ungueltigen Image-Tag setzen, Events lesen und manuell zurueckrollen.
5. HPA-Grenze von drei auf vier erhoehen und den begrenzten Lasttest ausfuehren.

Das Arbeitsblatt enthaelt die einzelnen Schritte. Der Lasttest erzeugt 150
Sekunden CPU-Last in einem Pod, danach endet er selbst. Der HPA nutzt den
Metrics Server aus k3s. Prometheus ist nicht seine CPU-Datenquelle.

## Erwarteter Endzustand

- Anwendung und PostgreSQL laufen weiter; `restaurant-pizza` hat eine Replica.
- Grafana zeigt Messwerte, der absichtlich erzeugte Rueckstau hat sich abgebaut.
- `betrieb-lab/lab-web` laeuft wieder mit dem gueltigen NGINX-Image und zwei
  bereiten Pods. Der HPA hat min=2, max=4 und CPU-Ziel 50 Prozent.
- Die eigene `labs/block-07/hpa.yaml` enthaelt ebenfalls maxReplicas=4.

Zur Fehlersuche zuerst `kubectl get pods`, `kubectl describe` und die Logs
im betroffenen Namespace pruefen. Weitere Demos aus der Referenzimplementation
sind nicht Teil dieses schlanken Materialpakets.
