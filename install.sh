#!/usr/bin/env sh
set -eu

SOURCE_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
TARGET_DIR=${1:-.}

if [ ! -f "$TARGET_DIR/deploy/overlays/block-05-messaging/kustomization.yaml" ]; then
  echo "Im Ziel fehlt der Projektstand aus Block 5 (deploy/overlays/block-05-messaging)." >&2
  exit 1
fi

for DIR in build cmd docs internal scripts platform/cloudnative-pg deploy/overlays/block-06-persistence; do
  mkdir -p "$TARGET_DIR/$DIR"
  cp -R "$SOURCE_DIR/$DIR/." "$TARGET_DIR/$DIR/"
done

cp "$SOURCE_DIR/go.mod" "$SOURCE_DIR/go.sum" "$TARGET_DIR/"

printf 'Block 6 wurde in %s installiert.\n' "$TARGET_DIR"
printf 'Naechster Schritt: CloudNativePG per Helm installieren, Images bauen und das Block-6-Overlay anwenden.\n'
