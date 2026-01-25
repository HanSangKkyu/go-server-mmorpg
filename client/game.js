const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');
const statusEl = document.getElementById('status');
const myIdEl = document.getElementById('myId');
const uiEl = document.getElementById('ui');

// Stats UI
const statsEl = document.createElement('div');
statsEl.innerHTML = 'HP: -/- | ATK: - | DEF: - | SPD: - | GOLD: 0';
uiEl.appendChild(statsEl);

const players = new Map();
const items = new Map();
const monsters = new Map();
const projectiles = new Map();

let myId = null;
let myStats = {};
let ws = null;

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.hostname;
    const wsUrl = `${protocol}//${host}:9000/ws`;

    console.log(`Connecting to ${wsUrl}`);
    statusEl.textContent = 'Connecting...';

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log('Connected');
        statusEl.textContent = 'Connected';
        statusEl.style.color = '#0f0';
    };

    ws.onclose = () => {
        console.log('Disconnected');
        statusEl.textContent = 'Disconnected';
        statusEl.style.color = '#f00';
        setTimeout(connect, 3000);
    };

    ws.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data);
            handleMessage(msg);
        } catch (e) {
            console.error('Invalid JSON:', event.data);
        }
    };
}

function handleMessage(msg) {
    switch (msg.type) {
        case 'SNAP':
            // Players
            const seenPlayers = new Set();
            if (msg.players) {
                msg.players.forEach(p => {
                    seenPlayers.add(p.id);
                    if (!players.has(p.id)) {
                        players.set(p.id, { x: p.x, y: p.y, color: getRandomColor() });
                    } else {
                        const existing = players.get(p.id);
                        if (p.id !== myId) {
                            existing.x = p.x;
                            existing.y = p.y;
                        }
                    }
                });
            }
            for (const [id] of players) { if (!seenPlayers.has(id)) players.delete(id); }

            // Monsters
            const seenMonsters = new Set();
            if (msg.monsters) {
                msg.monsters.forEach(m => {
                    seenMonsters.add(m.id);
                    if (!monsters.has(m.id)) {
                        monsters.set(m.id, { x: m.x, y: m.y, type: m.type, hp: m.hp, maxHp: m.max_hp });
                    } else {
                        const existing = monsters.get(m.id);
                        existing.x = m.x;
                        existing.y = m.y;
                        existing.type = m.type;
                        existing.hp = m.hp;
                        existing.maxHp = m.max_hp;
                    }
                });
            }
            for (const [id] of monsters) { if (!seenMonsters.has(id)) monsters.delete(id); }

            // Projectiles
            const seenProjs = new Set();
            if (msg.projectiles) {
                msg.projectiles.forEach(p => {
                    seenProjs.add(p.id);
                    if (!projectiles.has(p.id)) {
                        projectiles.set(p.id, { x: p.x, y: p.y });
                    } else {
                        const existing = projectiles.get(p.id);
                        existing.x = p.x;
                        existing.y = p.y;
                    }
                });
            }
            for (const [id] of projectiles) { if (!seenProjs.has(id)) projectiles.delete(id); }
            break;

        case 'ITEM_SPAWN':
            items.set(msg.id, { x: msg.x, y: msg.y });
            break;

        case 'ITEM_REMOVE':
            items.delete(msg.id);
            break;

        case 'GOLD_UPDATE':
            if (myStats) {
                myStats.gold = msg.amount;
                updateStatsUI();
            }
            break;

        case 'WELCOME':
            myId = msg.id;
            myIdEl.textContent = myId;
            myStats = {
                hp: msg.hp,
                maxHp: msg.max_hp,
                atk: msg.attack,
                def: msg.defense,
                speed: msg.speed,
                gold: msg.gold
            };
            updateStatsUI();
            break;

        case 'LEAVE':
            players.delete(msg.id);
            break;
    }
}

function updateStatsUI() {
    statsEl.textContent = `HP: ${myStats.hp}/${myStats.maxHp} | ATK: ${myStats.atk} | DEF: ${myStats.def} | SPD: ${myStats.speed} | GOLD: ${myStats.gold}`;
}

const keys = {};

window.addEventListener('keydown', (e) => {
    keys[e.key] = true;
});

window.addEventListener('keyup', (e) => {
    keys[e.key] = false;
});

function update() {
    if (!myId || !ws || ws.readyState !== WebSocket.OPEN) return;

    const me = players.get(myId);
    if (me) {
        let moved = false;
        let newX = me.x;
        let newY = me.y;
        const speed = myStats.speed || 5;

        if (keys['ArrowUp'] || keys['w']) { newY -= speed; moved = true; }
        if (keys['ArrowDown'] || keys['s']) { newY += speed; moved = true; }
        if (keys['ArrowLeft'] || keys['a']) { newX -= speed; moved = true; }
        if (keys['ArrowRight'] || keys['d']) { newX += speed; moved = true; }

        if (newX < 0) newX = 0;
        if (newY < 0) newY = 0;
        if (newX > canvas.width) newX = canvas.width;
        if (newY > canvas.height) newY = canvas.height;

        if (moved) {
            me.x = newX;
            me.y = newY;
            // Send JSON Move
            const packet = JSON.stringify({ type: 'MOVE', x: newX, y: newY });
            ws.send(packet);
        }
    }
}

function draw() {
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    // Items (Gold Squares)
    ctx.fillStyle = '#ffd700';
    items.forEach((item) => {
        ctx.fillRect(item.x - 5, item.y - 5, 10, 10);
    });

    // Monsters (Colored Squares based on Type)
    monsters.forEach((m) => {
        // 0: Water (Blue), 1: Fire (Red), 2: Grass (Green)
        if (m.type === 0) ctx.fillStyle = 'blue';
        else if (m.type === 1) ctx.fillStyle = 'red';
        else if (m.type === 2) ctx.fillStyle = 'green';

        ctx.fillRect(m.x - 10, m.y - 10, 20, 20);

        // Health Bar
        if (m.maxHp > 0) {
            const width = 30;
            const height = 4;
            const hpPct = Math.max(0, m.hp / m.maxHp);

            ctx.fillStyle = '#333';
            ctx.fillRect(m.x - 15, m.y - 20, width, height);

            ctx.fillStyle = '#f00';
            ctx.fillRect(m.x - 15, m.y - 20, width * hpPct, height);
        }
    });

    // Projectiles (Yellow Dots)
    ctx.fillStyle = '#ff0';
    projectiles.forEach((p) => {
        ctx.beginPath();
        ctx.arc(p.x, p.y, 3, 0, Math.PI * 2);
        ctx.fill();
    });

    // Players (Circles)
    players.forEach((p, id) => {
        ctx.fillStyle = id === myId ? '#ff0' : (p.color || '#fff');

        ctx.beginPath();
        ctx.arc(p.x, p.y, 10, 0, Math.PI * 2);
        ctx.fill();

        ctx.fillStyle = '#fff';
        ctx.font = '10px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(`P${id}`, p.x, p.y - 15);
    });
}

function loop() {
    update();
    draw();
    requestAnimationFrame(loop);
}

function getRandomColor() {
    const letters = '0123456789ABCDEF';
    let color = '#';
    for (let i = 0; i < 6; i++) {
        color += letters[Math.floor(Math.random() * 16)];
    }
    return color;
}

connect();
loop();
