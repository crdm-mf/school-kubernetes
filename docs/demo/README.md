# Nachweis der finalen Demo

Aufgenommen am 2026-09-14 nach vollständigem Durchlauf von
[README.md](../../README.md) (Cluster-Aufbau, Operatoren, Image-Build,
`sh scripts/build-images.sh`, `sh scripts/load-images.sh`, `kubectl apply -k
deploy/overlays/block-07-observability`, `sh scripts/smoke-test.sh`,
`sh scripts/failure-demo.sh`).

## Screenshots

| Datei | Zeigt |
|---|---|
| [01-dashboard.png](01-dashboard.png) | Dispatch-City-Dashboard, Stadtansicht mit laufender Simulation, Event-Stream, aktiven Bestellungen und Live-Status "14/14 Pods". |
| [07-dashboard-system.png](07-dashboard-system.png) | Systemansicht des Dashboards mit allen 8 Workloads (Dashboard 2/2, Control API 2/2, RabbitMQ 1/1, Customer/Courier Simulator 1/1, Restaurant Workers 3/3, Order Worker 2/2, CNPG Cluster 2/2) inklusive HTTP/SSE-, food.events- und SQL-Projection-Kanten. |
| [08-dashboard-system-post-failure.png](08-dashboard-system-post-failure.png) | Dieselbe Ansicht nach `scripts/failure-demo.sh`: Order-Worker auf 4/4 skaliert, `restaurant-pizza` neu erzeugt, System weiterhin ready. |
| [02-rabbitmq-overview.png](02-rabbitmq-overview.png) | RabbitMQ-Management-Overview mit Queued-Messages- und Publish/Deliver-Rates, Cluster-Info und 9 Queues / 10 Exchanges. |
| [03-rabbitmq-queues.png](03-rabbitmq-queues.png) | Liste der neun Queues (`courier-dispatch`, `food.dead`, `live.control-api-*`, `order-projection`, `restaurant.*`, `simulation-control.*`) mit DLX-Features und Message-Rates. |
| [04-rabbitmq-exchanges.png](04-rabbitmq-exchanges.png) | Exchange-Liste inklusive Topic-Exchange `food.events` und DLX `food.dlx`. |
| [05-grafana-dashboards.png](05-grafana-dashboards.png) | Grafana-Dashboards-Übersicht mit dem projektspezifischen Dashboard "Dispatch City - Betrieb" und den mitgelieferten Kubernetes-/CoreDNS-Boards. |
| [06-grafana-dispatch-city.png](06-grafana-dispatch-city.png) | Dashboard "Dispatch City - Betrieb" mit den fünf geforderten Panels: Offene Bestellungen, Gelieferte Bestellungen, Pizza-Worker bereit, Wartende Nachrichten je Restaurant, Verarbeitete Events pro Sekunde. |

## Logs

| Datei | Inhalt |
|---|---|
| [smoke-test.log](smoke-test.log) | Ausgabe von `scripts/smoke-test.sh` prüft `/health/ready`, Snapshot (`mode == distributed`, 3 Restaurants), manuellen Order-Trigger sowie Availability aller Deployments, StatefulSets und des CloudNativePG-Clusters. |
| [cluster-state.log](cluster-state.log) | `kubectl get pods,svc,statefulset,deployment,cluster,ingress -o wide` unmittelbar nach dem Deploy. |
| [rabbitmq-state.log](rabbitmq-state.log) | `rabbitmqctl list_exchanges|list_queues|list_bindings` inklusive `food.events`-Topic-Exchange, `food.dlx` und aller Restaurant-, Order-, Courier-Dispatch-, Simulation-Control- und Live-Queues. |
| [cnpg-cluster.log](cnpg-cluster.log) | Zustand des CloudNativePG-Clusters `food-delivery-db` (Primary `food-delivery-db-1`, ready-Instances 2/2, Image `postgresql:18.4-system-trixie`). |
| [failure-demo.log](failure-demo.log) | Ausgabe von `scripts/failure-demo.sh`: Löschen eines Restaurant-Pods, Rollout, Skalierung des Order-Workers auf 4 Replicas. |
| [post-demo-state.log](post-demo-state.log) | Zustand nach Failure-Demo: Order-Worker 4/4, Pizza-Pod neu, ServiceMonitors und PodMonitor aktiv, `courier-dispatch`-Backlog von 12 Messages, `order-projection` mit 4 Consumern. |

## Reproduktion

```bash
k3d cluster create delivery-lab --api-port 6550 -p "8080:80@loadbalancer" --agents 1
sh platform/cloudnative-pg/install.sh
sh platform/monitoring/install.sh
docker build -t food-delivery-dashboard:local apps/dashboard
sh scripts/build-images.sh
docker build -t food-delivery-cluster-observer:local --build-arg SERVICE=cluster-observer -f build/go-service.Dockerfile .
sh scripts/load-images.sh
k3d image import -c delivery-lab food-delivery-cluster-observer:local
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-07-observability
kubectl --context k3d-delivery-lab -n food-delivery wait --for=condition=Ready cluster/food-delivery-db --timeout=8m
kubectl --context k3d-delivery-lab -n food-delivery wait --for=condition=Available deployment --all --timeout=8m
sh scripts/smoke-test.sh
sh scripts/failure-demo.sh
```

Port-Forwards für Dashboards während der Aufnahme:

```bash
kubectl --context k3d-delivery-lab -n food-delivery port-forward svc/rabbitmq 15672:15672
kubectl --context k3d-delivery-lab -n monitoring port-forward svc/monitoring-grafana 3000:80
```

Grafana-Login `admin` / `delivery`, RabbitMQ-Login `delivery` / `delivery`
gemäss Manifesten unter `deploy/overlays/block-05-messaging/rabbitmq.yaml`
und `platform/monitoring/values-light.yaml`.
