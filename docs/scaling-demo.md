# Skalierungsdemo

Die fachlichen Entitäten werden bewusst mit nativen Kubernetes-Ressourcen skaliert. Das Dashboard liest keine
simulierten Slider-Werte, sondern die echten `ready`- und `desired`-Zahlen des Clusters.

```bash
./scripts/scale-city.sh couriers 6
./scripts/scale-city.sh customers 5
./scripts/scale-city.sh restaurant-pizza 3
```

Direkt mit `kubectl`:

```bash
kubectl --context k3d-delivery-lab -n food-delivery scale statefulset/courier-simulator --replicas=6
kubectl --context k3d-delivery-lab -n food-delivery scale statefulset/customer-simulator --replicas=5
kubectl --context k3d-delivery-lab -n food-delivery scale deployment/restaurant-pizza --replicas=3
```

## Beobachtung

1. In der Systemansicht steigt zuerst `desired`, danach `ready`.
2. Neue Customer- und Courier-Pods erhalten stabile Ordinalidentitäten und erscheinen als neue Figuren.
3. Restaurant-Replikas erscheinen als zusätzliche Küchenmodule am bestehenden Standort.
4. Beim Herunterskalieren verschwinden entfernte Kuriere mit ihren Pods. Eine noch nicht bestätigte Dispatch-Nachricht
   wird von RabbitMQ erneut zugestellt. Kunden mit einer offenen Bestellung bleiben bis zur Zustellung als `draining`
   sichtbar.
5. Der Event Stream zeigt Registrierung, Zuweisung, Fahrt zum Restaurant, Abholung, Fahrt zum Kunden und Zustellung.

Wenn die optionale Tracing-Erweiterung aktiviert ist, kann bei einer Bestellung aus der Dashboard-Bestellliste direkt
der zugehörige Trace in Grafana Explore geöffnet werden. So lässt sich ein beobachteter Rückstau mit der tatsächlichen
Verarbeitungsdauer einer einzelnen Order vergleichen. Die Suche verwendet die vollständige fachliche `correlation_id`,
nicht die gekürzte Anzeige-ID. Details stehen in [docs/tracing.md](tracing.md).
