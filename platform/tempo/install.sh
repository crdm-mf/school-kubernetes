#!/usr/bin/env sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
CONTEXT=${CONTEXT:-k3d-delivery-lab}
CHART_VERSION=${CHART_VERSION:-3.0.0}

helm upgrade --install tempo oci://ghcr.io/grafana-community/helm-charts/tempo \
  --kube-context "${CONTEXT}" \
  --namespace monitoring \
  --create-namespace \
  --version "${CHART_VERSION}" \
  --values "${ROOT}/values-light.yaml" \
  --wait \
  --timeout 8m

printf 'Tempo Chart %s ist im Kontext %s bereit.\n' "${CHART_VERSION}" "${CONTEXT}"
