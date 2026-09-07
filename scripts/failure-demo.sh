#!/usr/bin/env sh
set -eu

CONTEXT=${CONTEXT:-k3d-delivery-lab}
NAMESPACE=food-delivery

printf 'Deleting one restaurant worker pod...\n'
kubectl --context "${CONTEXT}" -n "${NAMESPACE}" delete pod -l food-delivery/entity-id=restaurant-pizza --wait=false
kubectl --context "${CONTEXT}" -n "${NAMESPACE}" rollout status deployment/restaurant-pizza --timeout=120s

printf 'Scaling order workers to four replicas...\n'
kubectl --context "${CONTEXT}" -n "${NAMESPACE}" scale deployment/order-worker --replicas=4
kubectl --context "${CONTEXT}" -n "${NAMESPACE}" rollout status deployment/order-worker --timeout=120s

printf 'The dashboard system view now reflects the changed workloads.\n'
