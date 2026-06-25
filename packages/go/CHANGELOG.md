# Changelog

All notable changes to the Oway Go SDK will be documented in this file.

## [0.3.0]

### Added

- `Client.GetCarrierTracking(ctx, identifier)` convenience wrapper returning the GPS tracking history (`CarrierTracking`: order number, last-updated time, and an ordered list of GPS points) for a carrier shipment.
- `Client.TrackShipmentStatusOnly(ctx, orderNumber)` for lightweight status-only polling that omits the live position estimate.

### Changed

- **Breaking:** `Client.TrackShipment` now includes the live position estimate (GPS center, uncertainty radius, last-event time, delay flags) by default, via the new `include=location` query the server accepts. Callers who want status-only polling should use `TrackShipmentStatusOnly`.
- **Breaking:** `CarrierTracking` is now an alias of the generated `TrackingResponse` (the GPS tracking-history shape the carrier tracking endpoint actually returns).
- **Breaking:** Regenerated from the latest sandbox spec. Email fields (for example `ShipperDispatch.Email` and address email fields) are now typed as `openapi_types.Email` (an alias of `string`) to reflect the `email` format in the spec. Callers passing a plain `string` literal will need to wrap it, for example `openapi_types.Email("dispatch@example.com")`.

## [0.2.4]

### Added

- `ShipperDispatch` type alias on the public SDK surface, so callers can populate the dispatch contact without dropping into the `client` package.
- `ShipmentRequest.ShipperDispatch *ShipperDispatch` field. Pass a `*ShipperDispatch` value to attach the operational dispatch contact (the desk responsible for answering questions about the load) on shipment create. At least one of `Email` or `Phone` must be set. Existing callers that leave it `nil` send the same JSON body as before.

## [0.2.3]

### Added

- `Appointments`, `AppointmentRequirement`, `AppointmentContact`, and `AppointmentChannel` type aliases on the public SDK surface, so callers can populate the appointment block without dropping into the `client` package. The underlying generated types landed in 0.2.2; this release exposes them through the hand-written wrapper.
- `ShipmentRequest.Appointments *Appointments` field. Pass an `*Appointments` value to attach pickup and/or delivery `AppointmentRequirement` payloads on shipment create. Existing callers that leave it `nil` send the same JSON body as before.

## [0.2.2]

### Added

- `OrderComponent` now carries optional per-line-item freight fields: `Description`, `NmfcCode`, `PackagingType`, `PieceCount`. All four are pointer types and pass through to the server unchanged when set. Existing callers that only set `PalletCount`, `PoundsWeight`, and `Dimensions` continue to work without changes.
- Generated client picks up everything else that landed on sandbox between v0.2.1 and v0.2.2, including the new appointment endpoints (`GetAppointment`, `UpsertAppointment`, `GetAppointmentPdf`, appointment-document upload/delete) and the optional `Idempotency-Key` header on shipment creation. Convenience wrappers for the appointment endpoints will arrive in a follow-up.

### Changed

- The generated `CreateShipmentWithResponse` signature now takes a `*CreateShipmentParams` (for the optional `Idempotency-Key` header). The SDK-level `Client.CreateShipment` wrapper still presents the same `(ctx, *ShipmentRequest)` shape and passes `nil` params internally; callers who need idempotency can drop down to `GeneratedClient()` for now.

## [0.2.1]

### Added

- `oway.PalletDims(length, width, height int32) *Dimensions` helper. Lets callers construct dimensions inline without taking the address of each int32.

### Fixed

- README examples no longer show the deprecated `PalletDimensions []int32` field; they now use the modern `Dimensions` object via `oway.PalletDims`.

### Note

- The `packages/go/v0.2.0` tag was published to the Go module proxy against an older revision before the v0.2.0 work landed on main. Because proxy content is immutable per version, the actual v0.2.0 changes ship under this v0.2.1 tag.

## [0.2.0]

### Added

- `Error` now exposes `Code` (from `ProblemDetail.reason`), `Message` (server `detail` or `title`), `RequestID`, `Violations []Violation`, and `RawBody []byte`.
- `Violation{Field, Reason}` type and `AsError(err) (*Error, bool)` helper for use with `errors.As`.
- `parseHTTPError(status, requestID, body)` helper for advanced callers that bypass the wrapper.
- Per-request company API key via context: `ctx = oway.WithCompanyAPIKey(ctx, key)` then call any method.
- Full-jitter exponential backoff retries on transient HTTP errors (`408/429/500/502/503/504`), driven off `Error.IsRetryable()`. Configurable via `Config.MaxRetries` (default 3; `-1` disables).
- `Config.Logger *slog.Logger` for structured debug/info/warn/error output. SDK is silent when nil.
- `golangci-lint` baseline configuration (gofmt, govet, errcheck, ineffassign, staticcheck, unused, misspell, revive).

### Changed

- **Breaking:** Dropped the `*ForCompany` method variants. Use `WithCompanyAPIKey(ctx, key)` on the context, then call the regular method.
- **Breaking:** `Client.GetClient()` renamed to `Client.GeneratedClient()` to reflect its purpose.
- **Breaking:** `GetQuoteByID` renamed to `GetQuote` for consistency with other resource accessors.
- Minimum Go version is now `1.22` (was `1.24.0`, which exceeded released versions).
- Request IDs are now 128-bit random hex from `crypto/rand`. Replaces the nanosecond-timestamp scheme, which collided under concurrency.

### Fixed

- `refreshToken` now surfaces JSON decode errors instead of silently returning an empty token + zero expiry.
- `refreshToken` validates `accessToken` is non-empty and `expiresIn` is positive before caching.
- Removed `fmt.Printf` debug write from `New`; debug output now flows through `Config.Logger`.
- Removed an unchecked `interface{}` to `string` type assertion in the auth transport.
- Wrapper error messages now include the server's `requestId` and parsed problem code rather than a bare HTTP status.

## [0.1.0] - 2026-02-19

### Added

- Initial release of Oway Go SDK
- M2M authentication with `ClientID` and `ClientSecret`
- Per-company API key support
- Clean type aliases: `oway.Quote`, `oway.Shipment` (no "External" prefix)
- Error types with `IsRetryable()` method
- Thread-safe token management with `sync.RWMutex`
- Context support for timeouts and cancellation
- Request ID generation for tracing
