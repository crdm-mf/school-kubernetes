# Block 6 - Helm, Operator und CloudNativePG

Dieser Baustein erweitert den Stand aus Block 5 um persistente Projektionen. CloudNativePG verwaltet einen PostgreSQL-Cluster mit Primary und Standby; der `order-worker` schreibt fachlichen Zustand und Idempotenzinformationen transaktional in die Datenbank.

## Verwendung im Kurs

- Integrationspaket fuer das fortlaufende Studierendenprojekt
- Startpunkt ist der funktionsfaehige Block-5-Stand
- Fuer den Kurs ist der reproduzierbare Release `v1.1.0` zu verwenden
- Docker Desktop, k3d, kubectl, Helm, Git und ein Browser genuegen
- Bash/macOS/WSL/Git Bash und Windows PowerShell werden unterstuetzt
- Die enthaltenen Zugangsdaten sind ausschliesslich fuer das lokale Kurs-Lab bestimmt

## Enthalten

- CloudNativePG Operator `1.30.0` als Helm Chart `0.29.0`
- `Cluster` Custom Resource mit zwei PostgreSQL-18.4-Instanzen auf getrennten Nodes
- je ein PVC pro Instanz sowie stabile Services fuer Schreib- und Lesezugriffe
- SQL-Migration fuer Projektionen, Eventhistorie und `processed_events`
- Migrationsjob und Secret-basierte Datenbankverbindung
- PostgreSQL-Repository fuer `order-worker` und `control-api`
- reproduzierbarer Idempotenz-Test mit gleicher `event_id`

## Integration

Das Installationsskript wird im Wurzelverzeichnis des bestehenden Projekts ausgefuehrt. Es ergaenzt den Code und das Overlay, ohne das Dashboard zu ersetzen.

```bash
../vsc-dispatch-city-06-persistence/install.sh .
./platform/cloudnative-pg/install.sh
./scripts/build-images.sh
CLUSTER=teko-k8s ./scripts/load-images.sh
kubectl --context k3d-teko-k8s apply -k deploy/overlays/block-06-persistence
```

```powershell
& "..\vsc-dispatch-city-06-persistence\install.ps1" -Target "."
./platform/cloudnative-pg/install.ps1
./scripts/build-images.ps1
./scripts/load-images.ps1 -Cluster teko-k8s
kubectl --context k3d-teko-k8s apply -k deploy/overlays/block-06-persistence
```

## Abnahme

- CloudNativePG meldet zwei bereite Instanzen.
- `food-delivery-db-rw` zeigt auf den aktuellen Primary.
- Eine Bestellung bleibt nach Neustarts von `order-worker` und `control-api` erhalten.
- Eine doppelt publizierte `event_id` wird nur einmal in `processed_events` gespeichert.
- Nach dem Loeschen des Primary-Pods wird ein Standby automatisch zum Primary.
