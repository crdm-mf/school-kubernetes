#!/usr/bin/env sh
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
CONTEXT=${CONTEXT:-k3d-delivery-lab}
CLUSTER=${CLUSTER:-delivery-lab}

CONTEXT="${CONTEXT}" "${ROOT}/platform/cloudnative-pg/install.sh"
CONTEXT="${CONTEXT}" "${ROOT}/platform/monitoring/install.sh"
TAG=local "${ROOT}/scripts/build-images.sh"
CLUSTER="${CLUSTER}" TAG=local "${ROOT}/scripts/load-images.sh"
if kubectl --context "${CONTEXT}" get namespace food-delivery >/dev/null 2>&1; then
  kubectl --context "${CONTEXT}" -n food-delivery delete deployment customer-simulator courier-simulator --ignore-not-found
fi
kubectl --context "${CONTEXT}" apply -k "${ROOT}/deploy/overlays/block-07-observability"
kubectl --context "${CONTEXT}" -n food-delivery wait --for=condition=Available deployment --all --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/rabbitmq --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/customer-simulator --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/courier-simulator --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery wait --for=condition=Ready cluster/food-delivery-db --timeout=8m
kubectl --context "${CONTEXT}" -n food-delivery get pods,svc,ingress
