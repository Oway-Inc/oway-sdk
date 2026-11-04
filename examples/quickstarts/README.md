# Shipper and carrier sandbox quickstarts

These examples create persistent sandbox records. Read [authentication](../../AUTH_GUIDE.md) and [environment setup](../../ENVIRONMENTS.md) first. Obtain a synthetic scenario and valid role credentials from Oway before running them; example addresses are placeholders, not a serviceability guarantee.

## Versions

- cURL with jq exercises the HTTP contract directly.
- Go shipper example uses the released `packages/go/v0.4.3` API, including `WithIdempotencyKey` and `GetDocument`.
- TypeScript examples require `@oway/sdk` **0.4.0 or later** for explicit creation idempotency, corrected document handling, and hosted POD helpers. npm 0.3.0 does not support these complete journeys.
- Carrier hosted-POD completion requires Go **0.4.4 or later** or TypeScript **0.4.0 or later**, plus the corresponding API capability. Confirm it is enabled in your environment before running.

## Shipper

Copy `quote.json` and `shipment.json` outside this checkout. Replace both address pairs with the assigned synthetic lane and keep cargo, quantities, freight class and accessorials consistent. Weight is per pallet; the sample uses one pallet to avoid ambiguous totals. The date is omitted for as-soon-as-possible service; for scheduled service use the current API's pickup-zone date semantics and a future fixture date.

Export these variables through your local secret manager or shell environment:

```text
OWAY_M2M_CLIENT_ID
OWAY_M2M_CLIENT_SECRET
OWAY_API_KEY
OWAY_QUOTE_FILE       absolute path to your quote JSON
OWAY_SHIPMENT_FILE    absolute path to your shipment JSON
OWAY_IDEMPOTENCY_KEY  unique stable key for this one creation
OWAY_STATE_FILE       new private file path for the exact create request and key
```

From the repository root:

```sh
bash examples/quickstarts/shipper.sh
# TypeScript, from this source checkout:
npm ci
npm run build --workspace=@oway/sdk
npx --yes tsx@4.20.6 packages/typescript/examples/quote-and-schedule.ts
# Go, from this source checkout:
cd packages/go
go run ./examples/shipper
```

The examples authenticate, quote, create, replay the same create key/body, confirm, track and retrieve a BOL. They require `CONFIRMED` in the response, not only HTTP success. A BOL may be temporarily unavailable; retrieve the existing shipment/document instead of creating another shipment. POD is a later carrier lifecycle document.

For a clean TypeScript consumer, create a new directory outside this checkout, run `npm init -y`, then `npm install @oway/sdk@0.4.0` and `npm install --save-dev tsx@4.20.6`. Copy `quickstart-config.ts` and the selected runner from `packages/typescript/examples` into that directory. Run `npx tsx quote-and-schedule.ts` or `npx tsx carrier-quickstart.ts` with the same environment variables. Install from the registry only after the required version is published; a workspace-source build is not a registry-consumer test.

For a clean Go consumer, copy `packages/go/examples/shipper/main.go` into a new module, run `go get github.com/Oway-Inc/oway-sdk/packages/go@v0.4.3`, and run it with the same environment. Use a fresh fixture and key for each independent complete run.

## Recovery

Do not restart an uncertain create from the beginning: a fresh quote changes the request body. Each runner saves the original body and key in `OWAY_STATE_FILE` before creating, with owner-only permissions, and refuses to overwrite an existing file. Retain it, then replay that exact creation or retrieve the known shipment. Extract the body with `jq .request "$OWAY_STATE_FILE"` and the key with `jq -r .idempotencyKey "$OWAY_STATE_FILE"`; do not request another quote for this replay. The state file contains synthetic shipment details, so keep it private. A mismatched key/body gets 409; an in-progress request needs reconciliation. The examples disable automatic SDK retries so you can see which step failed.

A 400 confirmation response with `invalid_fields` requires correcting the reported fields. A 422 coverage/load refusal requires changing the shipment or account configuration, not repeated requests. For a safe validation exercise, omit a required create field in the synthetic scenario, inspect the refusal, restore it, and use a fresh key.

Never put credentials, token responses, signed document URLs or real customer data into bug reports. Record the operation, timestamp, HTTP status/reason and request identifier. Contact support to set up or reset your sandbox scenario.

## Carrier hosted-POD journey

Obtain a synthetic pending offer, a small valid synthetic signed PDF, and the matching scenario coordinates. Copy `location.json` and replace its coordinates. Set the same credential variables as above using a **carrier** key, then set:

```text
OWAY_OFFER_ID        exact offer assigned to the test carrier
OWAY_POD_FILE        absolute path to your synthetic PDF (at most 10 MiB)
OWAY_LOCATION_FILE  absolute path to scenario coordinates JSON
OWAY_STATE_FILE     new private file path for the returned document identity
OWAY_STEP           prepare
```

Run the TypeScript carrier example from a built checkout:

```sh
npx --yes tsx@4.20.6 packages/typescript/examples/carrier-quickstart.ts
```

It retrieves the assigned offer, accepts it, confirms pickup, reports location, requests a storage URL, PUTs the PDF, and completes the upload. It saves the document identity and metadata before the PUT; it does not save tokens or signed URLs. The state file must not already exist. To submit a corrected POD after completion, register a new document.

**Review is a real prerequisite.** Completion creates a PENDING POD. An authorized reviewer must approve it with a verified recipient signature before you can confirm delivery with the document. Then set `OWAY_STEP=deliver` and rerun. The example reads shipment state first because review can already have triggered delivery. If still picked up, it submits `proof_of_delivery_document_id`. Never send both this field and `proof_of_delivery_url`, and never use a PUT URL as a POD URL.

After an uncertain completion, set `OWAY_STEP=confirm-upload` to retry the saved ID and metadata. If the initial storage PUT failed, request a new upload, save its new identity, and upload before confirmation; a saved ID alone does not mean bytes arrived. Do not reuse `prepare` blindly after an uncertain acceptance or delivery: retrieve the existing shipment and reconcile its state first.

A `409 pod_review_required` means review/signature prerequisites are unmet, not that a retry delay will fix it. A denied or archived document is not retrievable through the uploaded-document endpoint. Register a new corrected POD when necessary. Completion currently supports PROOF_OF_DELIVERY, not SIGNATURE uploads. Retrieval returns a fresh `download_url` with `expires_at`; persist the document ID rather than the URL.

Equivalent carrier runners, with the same variables and steps:

```sh
bash examples/quickstarts/carrier.sh
# Or, from packages/go:
go run ./examples/carrier
```

Each language run needs its own offer and state file. To recover from a failed storage PUT, retain the old state file, choose a **new** state-file path, and run `prepare` with the **same** offer ID. The runner observes PICKED_UP and does not repeat pickup. To recover from an uncertain successful completion, use `confirm-upload` with the original state file instead.

For a clean Go carrier consumer, copy `packages/go/examples/carrier/main.go` to a new module and run `go get github.com/Oway-Inc/oway-sdk/packages/go@v0.4.4`, then `go run .` with the same variables.
