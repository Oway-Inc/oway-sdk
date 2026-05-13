# Changelog

All notable changes to the Oway Go SDK will be documented in this file.

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
