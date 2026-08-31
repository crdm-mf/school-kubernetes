# Block 6 - CloudNativePG und Persistenz

Dieser Baustein ersetzt die fluechtige Projektion durch PostgreSQL. CloudNativePG verwaltet zwei Datenbankinstanzen, Rollenwechsel und den schreibbaren Service; der `order-worker` persistiert idempotent.

## Verwendung im Kurs

- Privates Integrationspaket fuer das bestehende Studierenden-Repository
- Kein neues Abgabe-Repository: Die Aenderungen werden in den fortlaufenden Projektstand uebernommen
- Fuer einen reproduzierbaren Stand ist der Release `v1.0.0` zu verwenden
- Die enthaltenen Zugangsdaten sind ausschliesslich fuer das lokale Kurs-Lab bestimmt

## Enthalten

- CNPG-Installation per Helm und `Cluster`-Ressource mit zwei Instanzen
- SQL-Migration fuer aktuelle Projektionen, Eventhistorie und `processed_events`
- Migrationsjob und Secret-basierte Datenbankverbindung
- PostgreSQL-Repository sowie DB-faehiger `order-worker` und `control-api`
- Persistente Customer-/Courier-Registrierung, Pickup-Phase und Strassenpositionen

## Arbeitsauftrag

1. Operator, Custom Resource und erzeugte Kubernetes-Ressourcen zuordnen.
2. CNPG installieren und die Persistenzkomponenten in Block 5 integrieren.
3. Schreib- und Lesezugriffe sowie die Rolle von `-rw`, `-ro` und `-r` Services pruefen.
4. Doppelte Event-IDs senden und die Idempotenz nachweisen.
5. Den Primary-Pod loeschen und Failover, Datenbestand sowie Anwendung beobachten.

```bash
CONTEXT=k3d-delivery-lab ./platform/cloudnative-pg/install.sh
kubectl --context k3d-delivery-lab apply -k deploy/overlays/block-06-persistence
kubectl --context k3d-delivery-lab -n food-delivery get cluster,pods
```

Abnahme: PostgreSQL meldet `2/2` Instanzen, Bestellungen und permanente Kurierpositionen ueberleben Worker-Neustarts und ein Primary-Failover fuehrt nicht zu verlorenem Projektzustand.
