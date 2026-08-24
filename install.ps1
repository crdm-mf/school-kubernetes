param(
    [string]$Target = "."
)

$ErrorActionPreference = "Stop"

$Source = Split-Path -Parent $MyInvocation.MyCommand.Path
$TargetPath = (Resolve-Path $Target).Path
$BaseManifest = Join-Path $TargetPath "deploy/base/kustomization.yaml"

if (-not (Test-Path $BaseManifest)) {
    throw "Im Ziel fehlt deploy/base/kustomization.yaml. Fuehren Sie das Skript im Projektstand aus Block 3 aus."
}

$OverlayTarget = Join-Path $TargetPath "deploy/overlays/block-04-ingress"
$RouteTarget = Join-Path $TargetPath "apps/dashboard/server/routes"
$PageTarget = Join-Path $TargetPath "apps/dashboard/pages/index.vue"
$BackupTarget = Join-Path $TargetPath ".block-backups"
$PageBackup = Join-Path $BackupTarget "dashboard-index.block-03.vue"
$LegacyBackup = Join-Path $TargetPath "apps/dashboard/pages/index.block-03.vue"

New-Item -ItemType Directory -Force -Path $OverlayTarget | Out-Null
New-Item -ItemType Directory -Force -Path $RouteTarget | Out-Null
New-Item -ItemType Directory -Force -Path $BackupTarget | Out-Null

Copy-Item -Recurse -Force (Join-Path $Source "deploy/overlays/block-04-ingress/*") $OverlayTarget
Copy-Item -Force (Join-Path $Source "apps/dashboard/server/routes/ui-instance.get.ts") $RouteTarget

if (Test-Path $LegacyBackup) {
    if (-not (Test-Path $PageBackup)) {
        Move-Item $LegacyBackup $PageBackup
    } else {
        Remove-Item $LegacyBackup
    }
}

if ((Test-Path $PageTarget) -and -not (Test-Path $PageBackup)) {
    Copy-Item $PageTarget $PageBackup
}
Copy-Item -Force (Join-Path $Source "apps/dashboard/pages/index.vue") $PageTarget

Write-Host "Block 4 wurde in $TargetPath installiert."
Write-Host "Naechster Schritt: Dashboard-Image bauen und in den Cluster teko-k8s importieren."
