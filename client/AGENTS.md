# CLIENT KNOWLEDGE BASE

## OVERVIEW
HTML5 Canvas frontend for the MMORPG. Handles real-time rendering, user input, and WebSocket synchronization with the Go server.

## STRUCTURE
- `index.html`: Game container and UI overlay.
- `game.js`: Main entry point. Contains the game loop, network handlers, and rendering logic.

## RENDERING
- **Loop**: Uses `requestAnimationFrame` for a smooth 60fps (ideally) draw cycle.
- **Canvas**: Direct 2D context manipulation.
- **Layers**:
    1. Background (Clear)
    2. Items (Gold squares)
    3. Monsters (Colored squares with health bars)
    4. Projectiles (Yellow arcs)
    5. Players (Circles with ID labels)

## NETWORK SYNC
- **WebSocket**: Connects to `ws://<host>:9000/ws`.
- **Input Prediction**: Client-side movement is applied immediately to the local player state before being sent to the server (`MOVE` packet).
- **Reconciliation**: Server snapshots (`SNAP`) overwrite non-local entity positions. Local player position is updated by the server but the client performs basic prediction to hide latency.
- **Event Handlers**:
    - `SNAP`: Full world state update (Players, Monsters, Projectiles).
    - `ITEM_SPAWN/REMOVE`: Incremental item updates.
    - `GOLD_UPDATE`: Local player stat synchronization.
    - `WELCOME`: Initial handshake, assigns `myId` and base stats.

## CONVENTIONS
- **State**: Managed via global `Map` objects (`players`, `items`, `monsters`, `projectiles`).
- **Colors**: Local player is always `#ff0` (Yellow). Others are assigned random hex codes on first sight.
- **Input**: Tracks `keys` object via `keydown`/`keyup` listeners. Supports Arrow keys and WASD.
