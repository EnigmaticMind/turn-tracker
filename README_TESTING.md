# Testing Guide

This document describes how to run tests for both the backend and frontend.

## Backend Testing

The backend uses Go's built-in `testing` package. Tests verify WebSocket message JSON formats and integration scenarios.

### Running Tests

```bash
cd backend
go test ./... -v
```

### Running Specific Test Packages

```bash
# Test create_room message format
go test ./handlers/createroom -v

# Test start_turn message format
go test ./handlers/startturn -v

# Run all tests with coverage
make test-coverage
```

### Test Structure

- **`test_helpers/`**: Utilities for creating test WebSocket servers and clients
- **`handlers/*/`**: Handler-specific tests for message format validation
- **`main_test.go`**: Integration tests for multiple clients and error handling

### What Tests Verify

- JSON message structure (correct fields, types, nullability)
- Message round-trip (send request → receive response)
- Multi-client scenarios (multiple clients in same room)
- Error message formats
- Turn state synchronization across clients

## Frontend Testing

The frontend uses Vitest for unit tests and Playwright for E2E tests.

### Running Tests

```bash
cd frontend

# Run unit tests
npm test

# Run tests in watch mode
npm test -- --watch

# Run tests with UI
npm run test:ui

# Run E2E tests (requires dev server running)
npm run test:e2e
```

### Test Structure

- **`src/lib/websocket/__tests__/`**: Unit tests for WebSocket message format validation
- **`e2e/`**: End-to-end tests using Playwright

### What Tests Verify

- JSON message structure validation
- Request/response format correctness
- Type safety for message data structures

### E2E Tests

E2E tests verify:

- Room creation flow
- Multi-player scenarios
- Turn changes and synchronization
- Real WebSocket communication

Note: E2E tests require both the backend server (`go run .` in `backend/`) and frontend dev server (`npm run dev` in `frontend/`) to be running, or they will use Playwright's `webServer` configuration.

## Continuous Integration

Tests should be run before deploying. Both backend and frontend have test commands that can be integrated into CI/CD pipelines.
