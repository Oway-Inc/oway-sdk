#!/usr/bin/env bash
set -euo pipefail
: "${OWAY_M2M_CLIENT_ID:?}" "${OWAY_M2M_CLIENT_SECRET:?}" "${OWAY_API_KEY:?}" "${OWAY_STATE_FILE:?}" "${OWAY_STEP:?}"
base=https://api.sandbox.oway.io
umask 077
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
jq -n '{clientId:env.OWAY_M2M_CLIENT_ID,clientSecret:env.OWAY_M2M_CLIENT_SECRET}' > "$work/auth.json"
curl --silent --show-error --fail -H 'Content-Type: application/json' --data-binary @"$work/auth.json" "$base/v1/auth/token" > "$work/token.json"
token=$(jq -er '.accessToken | select(length > 0)' "$work/token.json")
api() {
 printf 'API operation: %s\n' "${!#}" >&2
 curl --silent --show-error --fail -H "Authorization: Bearer $token" -H "x-oway-api-key: $OWAY_API_KEY" -H 'Content-Type: application/json' "$@"; }
if [ "$OWAY_STEP" = prepare ]; then
 : "${OWAY_OFFER_ID:?}" "${OWAY_POD_FILE:?}" "${OWAY_LOCATION_FILE:?}"
 test ! -e "$OWAY_STATE_FILE"
 size=$(wc -c < "$OWAY_POD_FILE")
 test "$size" -gt 0 && test "$size" -le 10485760
 offer=$(jq -nr --arg id "$OWAY_OFFER_ID" '$id|@uri')
 api "$base/v1/carrier/offers/$offer" > "$work/offer.json"
 if [ "$(jq -r '.status' "$work/offer.json")" = PENDING ]; then
  api -X PUT --data '{}' "$base/v1/carrier/offers/$offer/accept" > "$work/offer.json"
 fi
 if [ "$(jq -r '.status' "$work/offer.json")" = ACCEPTED ]; then
  jq -n --arg now "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '{pickup_time:$now}' > "$work/pickup.json"
  api -X PUT --data-binary @"$work/pickup.json" "$base/v1/carrier/shipments/$offer/pickup" > "$work/offer.json"
 fi
 jq -e '.status == "PICKED_UP"' "$work/offer.json" > /dev/null
 jq --arg now "$(date -u +%Y-%m-%dT%H:%M:%SZ)" '. + {timestamp:$now}' "$OWAY_LOCATION_FILE" > "$work/location.json"
 api --data-binary @"$work/location.json" "$base/v1/carrier/shipments/$offer/location" > "$work/location-result.json"
 jq -n --arg filename "$(basename "$OWAY_POD_FILE")" '{document_type:"PROOF_OF_DELIVERY",content_type:"application/pdf",file_name:$filename}' > "$work/metadata.json"
 api --data-binary @"$work/metadata.json" "$base/v1/carrier/shipments/$offer/upload" > "$work/upload.json"
 document=$(jq -er '.document_id|select(length>0)' "$work/upload.json")
 upload=$(jq -er '.upload_url|select(startswith("https://"))' "$work/upload.json")
 jq -e '.content_type == "application/pdf"' "$work/upload.json" > /dev/null
 jq -n --arg offer "$OWAY_OFFER_ID" --arg document "$document" --slurpfile metadata "$work/metadata.json" '{offerId:$offer,documentId:$document,metadata:$metadata[0]}' > "$work/state.json"
 (set -o noclobber; cat "$work/state.json" > "$OWAY_STATE_FILE")
 printf 'Upload POD bytes\n'
 curl --silent --show-error --fail --proto '=https' -X PUT -H 'Content-Type: application/pdf' --data-binary @"$OWAY_POD_FILE" "$upload" > /dev/null
 api -X PUT --data-binary @"$work/metadata.json" "$base/v1/carrier/shipments/$offer/uploads/$document/confirm" > "$work/completed.json"
 jq -e --arg id "$document" '.document_id == $id' "$work/completed.json" > /dev/null
 printf 'POD registered. Request authorized review, then run OWAY_STEP=deliver.\n'
else
 offer=$(jq -er '.offerId|@uri' "$OWAY_STATE_FILE")
 document=$(jq -er '.documentId|@uri' "$OWAY_STATE_FILE")
 if [ "$OWAY_STEP" = confirm-upload ]; then
  jq '.metadata' "$OWAY_STATE_FILE" > "$work/metadata.json"
  api -X PUT --data-binary @"$work/metadata.json" "$base/v1/carrier/shipments/$offer/uploads/$document/confirm" > "$work/completed.json"
 elif [ "$OWAY_STEP" = deliver ]; then
  api "$base/v1/carrier/shipments/$offer/documents/$document" > "$work/document.json"
  jq -e '.review_status == "APPROVED"' "$work/document.json" > /dev/null
  api "$base/v1/carrier/shipments/$offer" > "$work/shipment.json"
  if [ "$(jq -r '.status' "$work/shipment.json")" != DELIVERED ]; then
   jq -e '.status == "PICKED_UP"' "$work/shipment.json" > /dev/null
   jq -n --arg now "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --arg id "$(jq -r '.documentId' "$OWAY_STATE_FILE")" '{delivery_time:$now,proof_of_delivery_document_id:$id}' > "$work/delivery.json"
   api -X PUT --data-binary @"$work/delivery.json" "$base/v1/carrier/shipments/$offer/deliver" > "$work/shipment.json"
  fi
  jq -e '.status == "DELIVERED"' "$work/shipment.json" > /dev/null
  api "$base/v1/carrier/shipments/$offer/documents/$document" > "$work/document.json"
  printf 'Sandbox shipment delivered; registered POD retrieved.\n'
 else
  printf 'OWAY_STEP must be prepare, confirm-upload or deliver.\n' >&2; exit 1
 fi
fi
