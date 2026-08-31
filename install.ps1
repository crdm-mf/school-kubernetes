param(
    [string]$Target = "."
)

$ErrorActionPreference = "Stop"
$Source = Split-Path -Parent $MyInvocation.MyCommand.Path
$TargetPath = (Resolve-Path $Target).Path
$Block5Manifest = Join-Path $TargetPath "deploy/overlays/block-05-messaging/kustomization.yaml"

if (-not (Test-Path $Block5Manifest)) {
    throw "Im Ziel fehlt der Projektstand aus Block 5 (deploy/overlays/block-05-messaging)."
}

$Directories = @(
    "build",
    "cmd",
    "docs",
    "internal",
    "scripts",
    "platform/cloudnative-pg",
    "deploy/overlays/block-06-persistence"
)

foreach ($Directory in $Directories) {
    $Destination = Join-Path $TargetPath $Directory
    New-Item -ItemType Directory -Force -Path $Destination | Out-Null
    Copy-Item -Recurse -Force (Join-Path $Source "$Directory/*") $Destination
}

Copy-Item -Force (Join-Path $Source "go.mod") $TargetPath
Copy-Item -Force (Join-Path $Source "go.sum") $TargetPath

Write-Host "Block 6 wurde in $TargetPath installiert."
Write-Host "Naechster Schritt: CloudNativePG per Helm installieren, Images bauen und das Block-6-Overlay anwenden."
