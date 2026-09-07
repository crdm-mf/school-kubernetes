#!/usr/bin/env sh
set -eu

BASE_URL=${BASE_URL:-http://localhost:8080}
CONTEXT=${CONTEXT:-k3d-delivery-lab}

curl --fail --silent "${BASE_URL}/health/ready" >/dev/null
curl --fail --silent "${BASE_URL}/api/v1/snapshot" | jq -e '.mode == "distributed" and (.restaurants | length) == 3' >/dev/null
curl --fail --silent --request POST "${BASE_URL}/api/v1/orders" >/dev/null
kubectl --context "${CONTEXT}" -n food-delivery wait --for=condition=Available deployment --all --timeout=180s
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/rabbitmq --timeout=180s
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/customer-simulator --timeout=180s
kubectl --context "${CONTEXT}" -n food-delivery rollout status statefulset/courier-simulator --timeout=180s
kubectl --context "${CONTEXT}" -n food-delivery wait --for=condition=Ready cluster/food-delivery-db --timeout=180s
printf 'Smoke test passed: %s\n' "${BASE_URL}"
