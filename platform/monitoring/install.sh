#!/usr/bin/env sh
set -eu

CONTEXT=${CONTEXT:-k3d-delivery-lab}
CHART_VERSION=${CHART_VERSION:-88.1.3}
ROOT=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
VALUES_FILE=${VALUES_FILE:-}

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --force-update
set -- --values "${ROOT}/values-light.yaml"
if [ -n "${VALUES_FILE}" ]; then
	set -- "$@" --values "${VALUES_FILE}"
fi
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  --kube-context "${CONTEXT}" \
  --namespace monitoring \
  --create-namespace \
  --version "${CHART_VERSION}" \
  "$@" \
  --wait \
  --timeout 8m
