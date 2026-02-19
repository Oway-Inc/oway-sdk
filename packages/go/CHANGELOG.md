# Changelog

All notable changes to the Oway Go SDK will be documented in this file.

## [Unreleased]

## [0.1.0] - 2026-02-19

### Added
- Initial release of Oway Go SDK
- M2M authentication with `ClientID` and `ClientSecret`
- Per-company API key support via `ForCompany` methods
- Clean method names using `ogen` generator (no "WithResponse" suffix)
- Type aliases: `oway.Quote`, `oway.Shipment` (no "External" prefix)
- Error types with `IsRetryable()` method
- Thread-safe token management with `sync.RWMutex`
- Double-check locking pattern (production-proven)
- Context support for timeouts and cancellation
- Request ID generation for tracing

### Documentation
- Complete API reference
- Authentication guide
- Multi-company integration examples

### Testing
- 13 tests with 92% pass rate
- Token management tests
- Error classification tests
- Concurrent request tests
