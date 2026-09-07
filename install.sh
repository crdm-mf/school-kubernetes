#!/usr/bin/env sh
set -eu
SOURCE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
TARGET=${1:-.}
test -f "$TARGET/deploy/overlays/block-06-persistence/kustomization.yaml" || {
  echo 'Im Ziel fehlt der Projektstand aus AB6.' >&2
  exit 1
}
for DIR in cmd/cluster-observer internal/cluster platform/monitoring deploy/overlays/block-07-observability labs/block-07; do
  mkdir -p "$TARGET/$DIR"
  cp -R "$SOURCE/$DIR/." "$TARGET/$DIR/"
done
echo 'AB7 integriert. Weiter mit platform/monitoring/start-course.sh.'
