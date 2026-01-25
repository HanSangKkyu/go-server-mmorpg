# PROJECT KNOWLEDGE BASE

**Generated:** 2026-01-25
**Language:** Go 1.21+

## OVERVIEW
Simple MMORPG server in Go with a raw WebSocket/TCP JSON protocol.
Uses `gorilla/websocket` for transport and an in-memory game state loop.

## STRUCTURE
```
/
├── cmd/server/       # Entry point (main.go)
├── internal/
│   ├── game/         # Core logic (State, Player, Protocol, Types)
│   └── network/      # Transport (WS Server, Handler)
└── client/           # HTML5/Canvas Frontend (No build step)
```

## WHERE TO LOOK
| Task | Location | Notes |
|------|----------|-------|
| **Game Loop** | `internal/game/game.go` | `Update()` runs on 33ms ticker |
| **Protocol** | `internal/game/protocol.go` | JSON structs (`MsgSnap`, `MsgMove`) |
| **Entities** | `internal/game/types.go` | `Item`, `Monster`, `Projectile` defs |
| **Handlers** | `internal/network/handler.go` | Routes client commands |
| **Frontend** | `client/game.js` | Canvas rendering & WS logic |

## CODE MAP

| Symbol | Type | Location | Role |
|--------|------|----------|------|
| `Game` | struct | `internal/game/game.go` | Central state container, holds all maps |
| `Player` | struct | `internal/game/player.go` | Client session & stats state |
| `WSServer` | struct | `internal/network/websocket.go` | HTTP/WS upgrader & loop |
| `MsgSnap` | struct | `internal/game/protocol.go` | World state snapshot packet |

## CONVENTIONS
- **Locking**: `Game` uses `sync.RWMutex`. **Write Lock** (`Lock()`) for state mutations (Move, Spawn). **Read Lock** (`RLock()`) for Snapshots (if read-only, but currently Update takes Lock).
- **Communication**: JSON messages. Client sends commands (`MsgMove`), Server broadcasts updates (`MsgSnap`).
- **Structure**: Standard Go Layout (`cmd`, `internal`).

## ANTI-PATTERNS (THIS PROJECT)
- **Do NOT** use `fmt.Printf` for production logs; use `log`.
- **Do NOT** split packet strings manually; use `encoding/json`.
- **Do NOT** modify `internal/game` state without holding `g.lock`.

## COMMANDS
```bash
# Run Server
go run cmd/server/main.go

# Test
go test ./...
```
