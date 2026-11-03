#!/usr/bin/env bash
set -euo pipefail
: "${OWAY_M2M_CLIENT_ID:?}" "${OWAY_M2M_CLIENT_SECRET:?}" "${OWAY_API_KEY:?}"
: "${OWAY_STATE_FILE:?}"
[[ ! -e "$OWAY_STATE_FILE" ]] || { echo "State file exists; reconcile its saved creation before another run." >&2; exit 1; }
: "${OWAY_QUOTE_FILE:?}" "${OWAY_SHIPMENT_FILE:?}" "${OWAY_IDEMPOTENCY_KEY:?}"
base=https://api.sandbox.oway.io
umask 077
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
jq -n '{clientId:env.OWAY_M2M_CLIENT_ID,clientSecret:env.OWAY_M2M_CLIENT_SECRET}' > "$work/auth.json"
curl --silent --show-error --fail -H 'Content-Type: application/json' --data-binary @"$work/auth.json" "$base/v1/auth/token" > "$work/token.json"
token=$(jq -er '.accessToken | select(length > 0)' "$work/token.json")
api() {
 printf 'API operation: %s\n' "${!#}" >&2

 curl --silent --show-error --fail -H "Authorization: Bearer $token" -H "x-oway-api-key: $OWAY_API_KEY" -H 'Content-Type: application/json' "$@"
}
api --data-binary @"$OWAY_QUOTE_FILE" "$base/v1/shipper/quote" > "$work/quote.json"
quote=$(jq -er '.id | select(length > 0)' "$work/quote.json")
jq --arg id "$quote" '. + {quoteId:$id}' "$OWAY_SHIPMENT_FILE" > "$work/create.json"
(set -o noclobber; jq -n --slurpfile request "$work/create.json" '{request:$request[0],idempotencyKey:env.OWAY_IDEMPOTENCY_KEY}' > "$OWAY_STATE_FILE")
api -H "Idempotency-Key: $OWAY_IDEMPOTENCY_KEY" --data-binary @"$work/create.json" "$base/v1/shipper/shipment" > "$work/shipment.json"
number=$(jq -er '.orderNumber | select(length > 0)' "$work/shipment.json")
printf 'Created sandbox shipment %s; retain its identifier before proceeding.\n' "$number"
api -H "Idempotency-Key: $OWAY_IDEMPOTENCY_KEY" --data-binary @"$work/create.json" "$base/v1/shipper/shipment" > "$work/replay.json"
jq -e --arg number "$number" '.orderNumber == $number' "$work/replay.json" > /dev/null
api -X PUT "$base/v1/shipper/shipment/$number/confirm" > "$work/confirmed.json"
jq -e '.orderStatus == "CONFIRMED"' "$work/confirmed.json" > /dev/null
api "$base/v1/shipper/shipment/$number/tracking" > "$work/tracking.json"
api "$base/v1/shipper/shipment/$number/document/BILL_OF_LADING" > "$work/document.json"
jq -e '.downloadLink | startswith("https://")' "$work/document.json" > /dev/null
printf 'Sandbox shipment %s: confirmed, tracking and BOL retrieved.\n' "$number"
