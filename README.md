# Go MMORPG Server


## agent 행동 가이드
- 소스코드 수정 후 'omo' tmux session에서 실행중인 server 재실행
- 내가 git push 하라 할 때만 push 해, 그전에는 하지마

A simple, real-time MMORPG game server written in Go with a vanilla HTML5 Canvas client.

![Project Status](https://img.shields.io/badge/status-active-success.svg)
![Go Version](https://img.shields.io/badge/go-1.21%2B-blue.svg)

## 🎮 Features

- **Real-time Multiplayer**: See other players move and interact in real-time.
- **Combat System**:
  - Auto-aim projectiles targeting the nearest monster.
  - Elemental Monsters (Water 💧, Fire 🔥, Grass 🌿).
  - Health bars and damage mechanics.
- **Economy**:
  - Monsters drop items (Gold) upon death.
  - Inventory/Gold tracking system.
- **Technical Highlights**:
  - Raw WebSocket transport with JSON protocol.
  - Server-side authoritative game loop (30Hz).
  - Concurrent player handling using Goroutines.
  - Thread-safe state management with Mutexes.

## 🚀 Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) (version 1.21 or higher)
- A modern web browser

### Running the Server

1. **Clone the repository**
   ```bash
   git clone https://github.com/HanSangKkyu/go-server-mmorpg.git
   cd go-server-mmorpg
   ```

2. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```
   The server will start listening on port `9000`.

3. **Play the game**
   Open your browser and navigate to:
   [http://localhost:9000](http://localhost:9000)

   *Tip: Open multiple tabs to test multiplayer functionality!*

## 🕹️ Controls

- **Movement**: Arrow Keys or `W` `A` `S` `D`
- **Combat**: Automatic (Your character auto-shoots nearby monsters)
- **Looting**: Walk over dropped gold squares to collect them

## 🏗️ Architecture

### Backend (`internal/`)
- **Game Loop**: Runs continuously at ~30 ticks per second. Updates positions, collisions, and broadcasts state snapshots.
- **Protocol**: Custom JSON-based protocol.
  - `SNAP`: Full world state (Players, Monsters, Projectiles).
  - `MOVE`: Client input.
  - `WELCOME`: Initial handshake with stats.
- **Networking**: Uses `gorilla/websocket` for persistent connections.

### Frontend (`client/`)
- **Rendering**: HTML5 Canvas with `requestAnimationFrame` for smooth 60FPS rendering.
- **Logic**: Client-side prediction for local movement to ensure responsiveness.

## 📂 Project Structure

```
/
├── cmd/server/       # Entry point (main.go)
├── internal/
│   ├── game/         # Core game logic (State, Entities, Physics)
│   └── network/      # Network layer (WebSockets, JSON Handling)
├── client/           # Frontend assets (HTML, JS, CSS)
└── go.mod            # Go module definition
```

## 📝 License

This project is open source. Feel free to fork and modify!
