param([string]$Context = 'k3d-delivery-lab', [string]$ChartVersion = '3.0.0')
$ErrorActionPreference = 'Stop'

helm upgrade --install tempo oci://ghcr.io/grafana-community/helm-charts/tempo --kube-context $Context --namespace monitoring --create-namespace --version $ChartVersion --values (Join-Path $PSScriptRoot 'values-light.yaml') --wait --timeout 8m
if ($LASTEXITCODE -ne 0) { throw 'Tempo-Installation fehlgeschlagen.' }

Write-Host "Tempo Chart $ChartVersion ist im Kontext $Context bereit."
