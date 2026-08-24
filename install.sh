#!/usr/bin/env sh
set -eu

SOURCE_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
TARGET_DIR=${1:-.}

if [ ! -f "$TARGET_DIR/deploy/base/kustomization.yaml" ]; then
  echo "Im Ziel fehlt deploy/base/kustomization.yaml. Fuehren Sie das Skript im Projektstand aus Block 3 aus." >&2
  exit 1
fi

mkdir -p "$TARGET_DIR/deploy/overlays/block-04-ingress"
mkdir -p "$TARGET_DIR/apps/dashboard/server/routes"
mkdir -p "$TARGET_DIR/.block-backups"

cp -R "$SOURCE_DIR/deploy/overlays/block-04-ingress/." "$TARGET_DIR/deploy/overlays/block-04-ingress/"
cp "$SOURCE_DIR/apps/dashboard/server/routes/ui-instance.get.ts" "$TARGET_DIR/apps/dashboard/server/routes/"

PAGE_BACKUP="$TARGET_DIR/.block-backups/dashboard-index.block-03.vue"
LEGACY_BACKUP="$TARGET_DIR/apps/dashboard/pages/index.block-03.vue"

if [ -f "$LEGACY_BACKUP" ]; then
  if [ ! -f "$PAGE_BACKUP" ]; then
    mv "$LEGACY_BACKUP" "$PAGE_BACKUP"
  else
    rm "$LEGACY_BACKUP"
  fi
fi

if [ -f "$TARGET_DIR/apps/dashboard/pages/index.vue" ] && [ ! -f "$PAGE_BACKUP" ]; then
  cp "$TARGET_DIR/apps/dashboard/pages/index.vue" "$PAGE_BACKUP"
fi
cp "$SOURCE_DIR/apps/dashboard/pages/index.vue" "$TARGET_DIR/apps/dashboard/pages/index.vue"

printf 'Block 4 wurde in %s installiert.\n' "$TARGET_DIR"
printf 'Naechster Schritt: Dashboard-Image bauen und in den Cluster teko-k8s importieren.\n'
