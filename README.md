# Block 4 - Ingress und Load Balancing

Dieser Baustein erweitert den eigenen Projektstand aus Block 3. Er liefert:

- ein Traefik-Ingress fuer einen gemeinsamen HTTP-Einstiegspunkt,
- zwei Dashboard-Replikas fuer sichtbares Load Balancing,
- den Endpunkt `/ui-instance` zur Anzeige des antwortenden Dashboard-Pods.

Die `control-api` bleibt bei einer Replik, weil ihr Zustand in Block 4 noch im Speicher liegt.

## 1. Baustein uebernehmen

Fuehren Sie das Installationsskript aus Ihrem eigenen Projekt-Repository aus. Der Pfad zum Baustein kann bei Ihnen abweichen.

PowerShell:

```powershell
..\vsc-dispatch-city-04-ingress\install.ps1 -Target .
```

macOS, Linux, Git Bash oder WSL:

```bash
../vsc-dispatch-city-04-ingress/install.sh .
```

Das Skript sichert die bisherige Dashboard-Seite einmalig unter `.block-backups/dashboard-index.block-03.vue`.

## 2. Dashboard neu bauen und deployen

PowerShell:

```powershell
docker build -t food-delivery-dashboard:local .\apps\dashboard
k3d image import -c teko-k8s food-delivery-dashboard:local
kubectl apply -k .\deploy\overlays\block-04-ingress
kubectl -n food-delivery rollout restart deployment/dashboard
kubectl -n food-delivery rollout status deployment/dashboard
```

macOS, Linux, Git Bash oder WSL:

```bash
docker build -t food-delivery-dashboard:local ./apps/dashboard
k3d image import -c teko-k8s food-delivery-dashboard:local
kubectl apply -k ./deploy/overlays/block-04-ingress
kubectl -n food-delivery rollout restart deployment/dashboard
kubectl -n food-delivery rollout status deployment/dashboard
```

## 3. Traefik lokal oeffnen

Der Cluster aus Block 2 besitzt keine feste Host-Port-Zuordnung. Deshalb wird der bereits installierte Traefik-Service lokal weitergeleitet:

```powershell
kubectl -n kube-system port-forward service/traefik 8080:80
```

Danach sind Dashboard und API ueber denselben Einstiegspunkt erreichbar:

- Dashboard: `http://localhost:8080/`
- Snapshot: `http://localhost:8080/api/v1/snapshot`
- Readiness: `http://localhost:8080/health/ready`

## 4. Load Balancing sichtbar machen

PowerShell:

```powershell
1..20 | ForEach-Object { (Invoke-RestMethod -Headers @{ Connection = "close" } http://localhost:8080/ui-instance).instance } | Sort-Object -Unique
```

macOS, Linux, Git Bash oder WSL:

```bash
for i in $(seq 1 20); do curl -s -H 'Connection: close' http://localhost:8080/ui-instance; echo; done | sort -u
```

Erwartetes Resultat: In der Ausgabe erscheinen die Namen beider Dashboard-Pods.
