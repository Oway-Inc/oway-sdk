# Changelog

All notable changes to the Oway TypeScript SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-11-03

### Added

- Runnable shipper and carrier quickstarts with persistent recovery state, business-state checks, and sandbox prerequisites.
- Carrier `requestUploadUrl`, `confirmUpload`, and `uploadedDocument` helpers, with typed upload metadata and optional `proof_of_delivery_document_id` for delivery after authorized POD review.
- `shipments.create(params, companyApiKey, { idempotencyKey })` sets the creation key; `CreateShipmentOptions` is exported.
- `pickupAccessorials` and `deliveryAccessorials` on `Quote`, with the `QuoteStopAccessorials` type. Omitted when the quote recorded none.
- Hazardous material declarations on quotes and shipments, with the `HazmatPaper`, `HazmatEntry`, `HazmatCertification`, `HazmatClassification` and `HazmatFinding` types. See `docs/hazmat-capture.md`.
- Optional per-component `freightClass`, shipment `instructions` and `referenceNumbers`, invoice line-item `nmfcCode`, carrier milestone reporting types, and quote `shipmentMode`.
- Optional address email, phone extension, country, inside-service and lumper-service fields; `customerReference` on quotes; carrier address country/phone extension; the `REJECTED` status; generated quote batch and quote-by-reference operations.

### Changed

- Shipment references allow up to 100 entries, each at most 128 characters. Supplied order and repeated values are preserved. Confirm API support before submitting more than ten references.
- New fields remain optional for existing callers. Address country defaults to US.

### Fixed

- Shipment document helpers send canonical document types and map `downloadLink` into the existing `{url}` result. Existing BOL and LABEL aliases remain supported.
- Creation retries reuse one idempotency key, including when the caller omits an explicit key. `OwayError.isRetryable()` recognizes `idempotency_key_in_progress`.
- Properties with API defaults, such as address country, remain optional in generated request types.
- Authentication and onboarding links explain the required company key and environment setup.
- Missing sandbox credentials explicitly skip integration tests; quickstart tests run against SDK source before building. Examples are typechecked by the existing lint command.

## [0.2.2]

### Added

- `OrderComponent` now carries optional per-line-item freight fields: `description`, `nmfcCode`, `packagingType`, `pieceCount`. All four are optional and pass through to the server unchanged when set. Existing callers that only set `palletCount`, `poundsWeight`, and `dimensions` continue to work without changes.
- Generated schema picks up everything else that landed on sandbox between v0.2.1 and v0.2.2, including the new appointment endpoints (`GET/PUT/DELETE /v1/shipper/shipment/{orderNumber}/appointment/{stop}` and the appointment-document + PDF variants) and the optional `Idempotency-Key` header on shipment creation.

## [0.2.1]

### Fixed

- README examples now use the modern `dimensions: { length, width, height }` object instead of the deprecated `palletDimensions: [l, w, h]` array.

## [0.2.0]

### Added

- Typed `OwayError` carrying the parsed `ProblemDetail`: `statusCode`, `code` (machine-readable reason), `message` (server detail or title), `requestId`, `violations[]`, and `rawBody`.
- `parseHttpError(status, requestId, body)` helper exported for advanced callers that consume responses directly.
- Public `ProblemDetail` and `Violation` type re-exports for users who need to read structured validation failures.
- Full-jitter exponential backoff on retries (was fixed delay), bounded at 30s.
- ESLint flat-config baseline with TypeScript rules.

### Changed

- **Breaking:** `OwayError` constructor now takes an options object (`new OwayError({ message, statusCode, code, requestId, violations, rawBody })`). Old positional form (`new OwayError(message, code, statusCode, requestId)`) is removed.
- `isRetryable()` is now consistent across the SDK and recognizes `408 Request Timeout` in addition to `429/500/502/503/504`.
- Request IDs use `crypto.randomUUID()` when available, falling back to a 128-bit random hex string. Replaces the hand-rolled UUID-like generator.

### Fixed

- Error message extraction now reads `ProblemDetail.detail` / `.title` instead of a non-existent `message` field.
- Error code extraction now reads `ProblemDetail.reason` instead of a non-existent `code` field.
- Removed dead `companyApiKey` field from `OwayConfig`; per-request override flows through the existing `companyApiKey` argument on resource methods.

## [0.1.0] - 2026-02-19

### Added

- Initial release of Oway TypeScript SDK
- M2M authentication with `clientId` and `clientSecret`
- Per-company API key support for multi-tenant integrations
- Resource-based API: `oway.quotes.create()`, `oway.shipments.create()`
- Clean type aliases: `Quote`, `Shipment`, `Tracking`
- Smart error classification with `OwayError.isRetryable()`
- Automatic retry with exponential backoff
- Request ID tracking for debugging
- Safe logging with credential sanitization
- Thread-safe token management
- Model Context Protocol (MCP) server for AI agents

### Documentation

- Complete API reference
- Authentication guide (M2M + API keys)
- AI agent integration guide
- Multi-company integration examples
