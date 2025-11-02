# Component Tests

This directory contains isolated unit tests for React components in the Turn Tracker application.

## Test Structure

Tests are organized by component, with each component having its own test file:

- `PlayerGrid.test.tsx` - Tests for the player grid display component
- `GameHome.test.tsx` - Tests for the game home screen
- `PlayerTurn.test.tsx` - Tests for the player turn screen
- `GameContainer.test.tsx` - Tests for the main game container
- `Options.test.tsx` - Tests for the options/settings modal

## Test Utilities

`testUtils.tsx` provides:

- `createMockWebSocketManager()` - Creates a mock WebSocket manager for testing
- `renderWithRouter()` - Renders components with mocked React Router context
- `resetMocks()` - Cleans up mocks between tests

## Mocking Strategy

All external dependencies are mocked:

- **React Router hooks**: `useLoaderData`, `useParams`, `useNavigate`, `useOutletContext`
- **WebSocket handlers**: `startTurn`, `endTurn`, `updateProfile`
- **Child components**: `LiquidAvatar`, `PlayerGrid`, `Options`
- **Browser APIs**: `navigator.clipboard`, `navigator.wakeLock`

## Running Tests

```bash
# Run all component tests
npm test -- src/components/__tests__

# Run specific test file
npm test -- src/components/__tests__/PlayerGrid.test.tsx

# Run in watch mode
npm test -- --watch src/components/__tests__
```

## Test Coverage

Tests cover:

- Component rendering with various props
- User interactions (clicks, form inputs)
- Async operations (WebSocket handlers)
- Timer-based updates (elapsed time display)
- Error handling
- Edge cases (null/undefined values, empty arrays, etc.)
