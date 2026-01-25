# INTERNAL/GAME: CORE DOMAIN LOGIC

## OVERVIEW
This package implements the authoritative game state and simulation logic. It manages the lifecycle of all entities (Players, Monsters, Items, Projectiles) and handles the high-frequency tick loop.

## STRUCTURE
- `game.go`: Central `Game` struct, tick loop, and state mutation methods.
- `player.go`: Player entity logic, movement, and connection abstraction.
- `types.go`: Data structures for `Monster`, `Item`, and `Projectile`.
- `protocol.go`: JSON message definitions for client-server communication.

## KEY MECHANICS

### Tick Loop (33ms)
The `Start()` method runs a ticker at ~30 FPS. Each tick triggers `Update()` which:
1. Updates projectile positions.
2. Processes player auto-shooting logic.
3. Performs collision detection.
4. Broadcasts a world snapshot (`MsgSnap`) to all connected players.

### State Management (Mutex)
- **`Game.lock (sync.RWMutex)`**: Protects all entity maps.
- **Write Lock**: Required for `Update()`, `AddPlayer()`, `RemovePlayer()`, and `SpawnMonster()`.
- **Read Lock**: Used for state inspection (though currently `Update` takes a full lock for simplicity).

### Entity Lifecycles
- **Spawn**: Monsters spawn via `monsterTicker` (1s). Items spawn upon monster death.
- **Remove**: Projectiles are removed when they exit boundaries or collide. Monsters are removed when HP <= 0. Players are removed on disconnect.

### Collision Detection
Simple distance-based checks (`dx*dx + dy*dy < radius*radius`) are performed in `checkCollisions()` to handle:
- Projectile vs Monster (Damage/Death).
- Player vs Item (Collection).

## CONVENTIONS
- **ID Generation**: Use `g.lastID++` patterns within a lock to ensure unique entity IDs.
- **Broadcasting**: Use `g.broadcastJSON()` for event-driven updates (e.g., `ITEM_SPAWN`).
- **Positioning**: All coordinates are `float64`. Boundary checks are hardcoded in `updateProjectiles`.
