#!/usr/bin/env sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
CONTEXT=${CONTEXT:-k3d-delivery-lab}

CONTEXT="${CONTEXT}" sh "${ROOT}/platform/tempo/install.sh"
CONTEXT="${CONTEXT}" VALUES_FILE="${ROOT}/platform/monitoring/values-tracing.yaml" sh "${ROOT}/platform/monitoring/install.sh"
kubectl --context "${CONTEXT}" apply -k "${ROOT}/deploy/overlays/optional-tracing"
kubectl --context "${CONTEXT}" -n food-delivery wait --for=condition=Available deployment --all --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/customer-simulator --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/courier-simulator --timeout=8m

printf 'Optionales Distributed Tracing ist im Kontext %s aktiviert.\n' "${CONTEXT}"
