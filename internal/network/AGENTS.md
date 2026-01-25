# TRANSPORT LAYER (internal/network)

## OVERVIEW
Handles raw transport protocols (WebSocket and TCP) and maps them to the game's `Connection` interface. This layer is responsible for the connection lifecycle, message framing, and initial command unmarshalling.

## CONNECTION FLOW
1. **Upgrade/Accept**: 
   - `WSServer.HandleWS` upgrades HTTP to WebSocket.
   - `Server.acceptLoop` accepts raw TCP connections.
2. **Registration**: 
   - New connections are wrapped in a protocol-specific struct (e.g., `WSConnection`).
   - `game.AddPlayer(conn)` is called to create a `Player` entity and start tracking.
3. **Read Loop**:
   - Each connection runs a dedicated goroutine reading messages.
   - WebSocket uses `conn.ReadMessage()`.
   - TCP uses `bufio.NewScanner()` for line-delimited packets.
4. **Teardown**:
   - On read error or disconnect, `game.RemovePlayer(id)` is called via `defer`.
   - Connections are closed explicitly to free resources.

## HANDLERS
The `HandleCommand` function in `handler.go` acts as the boundary between raw transport and game logic:

- **JSON Unmarshalling**: 
  - Incoming strings are trimmed and checked for length.
  - Messages are first attempted to be unmarshalled into `game.MsgMove`.
  - Type-based routing (e.g., `move.Type == "MOVE"`) determines which `Player` method to invoke.
- **Concurrency**:
  - `WSConnection` uses a `sync.Mutex` on `Write` to ensure thread-safe broadcasts from the game loop.
  - Read loops are single-threaded per player, but multiple players read in parallel.

## KEY TYPES
| Symbol | Role |
|--------|------|
| `WSConnection` | Thread-safe wrapper for `websocket.Conn` |
| `WSServer` | HTTP handler for WebSocket entry point |
| `Server` | TCP listener and accept loop manager |
| `HandleCommand` | Entry point for protocol unmarshalling |
