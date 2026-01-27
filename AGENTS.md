# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-27
**Language:** Go 1.21+

## 1. OVERVIEW
Simple MMORPG server in Go with a raw WebSocket/TCP JSON protocol.
Uses `gorilla/websocket` for transport and an in-memory game state loop.
The client is a vanilla HTML5 Canvas application with no build step.

## 2. COMMANDS & ENVIRONMENT
- **Run Server**: `go run cmd/server/main.go`
- **Run Tests**: `go test ./...`
- **Run Single Test**: `go test -v -run TestName ./path/to/pkg`
- **Lint**: `go vet ./...` (Standard Go tools)
- **Format**: `go fmt ./...` (Run before committing)

## 3. PROJECT STRUCTURE
```
/
├── cmd/server/       # Entry point (main.go)
├── internal/
│   ├── game/         # Core Domain Logic (State, Player, Entities)
│   └── network/      # Network Layer (WebSockets, JSON Handling)
└── client/           # Frontend assets (HTML, JS, CSS)
```

## 4. CODE STYLE & CONVENTIONS

### General
- **Idiomatic Go**: Follow standard Go conventions (Effective Go).
- **Formatting**: ALWAYS run `go fmt`.
- **Imports**: Group imports: Standard Library, then 3rd Party, then Local.

### Naming
- **Structs/Interfaces**: PascalCase (exported), camelCase (private).
- **Variables**: Short but descriptive (e.g., `p` for player in small scopes, `player` in larger).
- **Constants**: PascalCase (e.g., `MonsterTypeFire`).

### Error Handling
- **Log, Don't Crash**: Use `log.Println` or `log.Printf`.
- **Fatal Errors**: Only use `log.Fatal` during startup (e.g., failed to bind port).
- **Return Errors**: Functions should return `error` as the last return value.

### Thread Safety (CRITICAL)
- **Global State**: `Game` struct holds all state.
- **Locking**:
  - Use `g.lock.Lock()` for ALL state mutations (Moving, Spawning, Equipping).
  - Use `g.lock.RLock()` for read-only snapshots (if applicable).
  - **Deadlock Prevention**: Never acquire a lock inside a function that is already called under a lock.

### Protocol
- **JSON**: All communication is JSON.
- **Struct Tags**: Use `json:"fieldName"` tags.
- **Messages**:
  - `Type`: String discriminator (e.g., "MOVE", "SNAP", "EQUIP").
  - `Payload`: Flattened fields in the struct.

## 5. ARCHITECTURE DETAILS

### Game Loop (`internal/game/game.go`)
- **Ticker**: Runs every 33ms (30 TPS).
- **Update()**:
  1. Lock State.
  2. Handle Physics/Logic.
  3. Unlock.
  4. Broadcast Snapshot.

### Networking (`internal/network/`)
- **WebSocket**: Using `gorilla/websocket`.
- **Handling**: Each client has a read pump (goroutine). Writes are synchronized.

## 6. FEATURE DEVELOPMENT GUIDE

### Adding New Entities
1. Define struct in `internal/game/types.go`.
2. Add to `Game` or `Map` struct.
3. Update `Update()` logic in `internal/game/game.go`.
4. Update `MsgSnap` in `internal/game/protocol.go`.

### Adding New Packets
1. Define `MsgXxx` struct in `internal/game/protocol.go`.
2. Handle in `internal/network/handler.go` (Server) or `client/game.js` (Client).

## 7. TESTING
- **Unit Tests**: Place `_test.go` files next to the code they test.
- **Mocking**: Use interfaces for network connections if needed.
