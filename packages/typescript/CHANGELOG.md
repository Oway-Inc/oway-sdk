# Changelog

All notable changes to the Oway TypeScript SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-19

### Added
- Initial release of Oway TypeScript SDK
- M2M authentication with `clientId` and `clientSecret`
- Per-company API key support for multi-tenant integrations
- Resource-based API: `oway.quotes.create()`, `oway.shipments.create()`
- Clean type aliases: `Quote`, `Shipment`, `Tracking` (no "External" prefix)
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

### Testing
- 31 unit tests with 97% pass rate
- Error handling tests
- Token management tests
- Concurrent request tests
