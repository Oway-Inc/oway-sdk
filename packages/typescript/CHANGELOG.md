# Changelog

All notable changes to the Oway TypeScript SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
