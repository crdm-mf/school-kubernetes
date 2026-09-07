#!/usr/bin/env sh
set -eu

CONTEXT=${CONTEXT:-k3d-delivery-lab}
NAMESPACE=${NAMESPACE:-food-delivery}
ENTITY=${1:-}
REPLICAS=${2:-}

if [ -z "${ENTITY}" ] || [ -z "${REPLICAS}" ]; then
  printf 'Usage: %s {couriers|customers|restaurants|restaurant-pizza|restaurant-bowl|restaurant-curry} REPLICAS\n' "$0" >&2
  exit 2
fi

case "${REPLICAS}" in
  *[!0-9]*|'')
    printf 'REPLICAS must be a non-negative integer.\n' >&2
    exit 2
    ;;
esac

case "${ENTITY}" in
  couriers)
    RESOURCE=statefulset/courier-simulator
    ;;
  customers)
    RESOURCE=statefulset/customer-simulator
    ;;
  restaurants)
    for restaurant in restaurant-pizza restaurant-bowl restaurant-curry; do
      kubectl --context "${CONTEXT}" -n "${NAMESPACE}" scale "deployment/${restaurant}" --replicas="${REPLICAS}"
      kubectl --context "${CONTEXT}" -n "${NAMESPACE}" rollout status "deployment/${restaurant}" --timeout=180s
    done
    printf 'Scaled all restaurant kitchens to %s pods.\n' "${REPLICAS}"
    exit 0
    ;;
  restaurant-pizza|restaurant-bowl|restaurant-curry)
    RESOURCE="deployment/${ENTITY}"
    ;;
  *)
    printf 'Unknown entity %s.\n' "${ENTITY}" >&2
    exit 2
    ;;
esac

kubectl --context "${CONTEXT}" -n "${NAMESPACE}" scale "${RESOURCE}" --replicas="${REPLICAS}"
kubectl --context "${CONTEXT}" -n "${NAMESPACE}" rollout status "${RESOURCE}" --timeout=180s
printf 'Scaled %s to %s pods. The city view updates within a few seconds.\n' "${ENTITY}" "${REPLICAS}"
