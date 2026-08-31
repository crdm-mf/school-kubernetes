#!/usr/bin/env sh
set -eu

CONTEXT=${CONTEXT:-k3d-delivery-lab}
CHART_VERSION=${CHART_VERSION:-0.29.0}

helm repo add cnpg https://cloudnative-pg.github.io/charts --force-update
helm upgrade --install cnpg cnpg/cloudnative-pg \
  --kube-context "${CONTEXT}" \
  --namespace cnpg-system \
  --create-namespace \
  --version "${CHART_VERSION}" \
  --wait \
  --timeout 5m
