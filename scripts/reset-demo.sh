#!/usr/bin/env sh
set -eu

BASE_URL=${BASE_URL:-http://localhost:8080}
curl --fail --silent --request POST "${BASE_URL}/api/v1/simulation/reset" >/dev/null
printf 'Simulation reset.\n'
