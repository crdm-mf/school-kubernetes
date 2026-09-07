# Event-Katalog

| Event | Producer | Consumer | Zweck |
|---|---|---|---|
| `customer.registered` | Customer Simulator | Order Worker, Control API | Stabilen Customer-Pod in der Stadt anmelden |
| `courier.registered` | Courier Simulator | Order Worker, Control API | Permanenten Courier-Pod in der Flotte anmelden |
| `order.created` | Customer Simulator, Control API | zuständiger Restaurant Worker, Order Worker | Neue Bestellung |
| `order.accepted` | Restaurant Worker | Courier Simulator, Order Worker | Bestellung angenommen |
| `order.rejected` | Restaurant Worker | Order Worker | Bestellung abgelehnt |
| `courier.assigned` | Courier Simulator | Order Worker | Kurier zugewiesen; Fahrt zum Restaurant beginnt |
| `courier.location.updated` | Courier Simulator | Order Worker | Aktuelle Kartenposition |
| `order.picked_up` | Courier Simulator | Order Worker | Abholung abgeschlossen; Fahrt zum Kunden beginnt |
| `order.delivered` | Courier Simulator | Order Worker | Lieferung abgeschlossen |
| `simulation.started` | Control API | Customer Simulator | Automatische Bestellungen starten |
| `simulation.paused` | Control API | Customer Simulator | Automatische Bestellungen pausieren |
| `simulation.reset` | Control API | alle Projektionen | Demo-Zustand leeren |

Alle Events verwenden `event_id`, `event_version`, Zeitstempel, `correlation_id`, `causation_id`, `source` und einen typisierten Payload. RabbitMQ liefert at-least-once; PostgreSQL verhindert doppelte Verarbeitung über `processed_events`.

Die Courier-Position liegt in jeder Phase auf dem gemeinsamen Strassengraphen. `progress` verwendet `0.0-0.45` fuer die Anfahrt zum Restaurant, `0.5` fuer die Abholung und `0.5-1.0` fuer die Fahrt zum Kunden.
