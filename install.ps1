param([string]$Target = '.')
$ErrorActionPreference = 'Stop'
$Destination = (Resolve-Path -LiteralPath $Target).Path
if (-not (Test-Path (Join-Path $Destination 'deploy/overlays/block-06-persistence/kustomization.yaml'))) {
    throw 'Im Ziel fehlt der Projektstand aus AB6.'
}
foreach ($Directory in @('cmd/cluster-observer', 'internal/cluster', 'platform/monitoring', 'deploy/overlays/block-07-observability', 'labs/block-07')) {
    $Path = Join-Path $Destination $Directory
    New-Item -ItemType Directory -Force -Path $Path | Out-Null
    Copy-Item -Recurse -Force (Join-Path $PSScriptRoot "$Directory/*") $Path
}
Write-Host 'AB7 integriert. Weiter mit platform/monitoring/start-course.ps1.'
