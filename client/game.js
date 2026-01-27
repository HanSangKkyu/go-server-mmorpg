const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');
const statusEl = document.getElementById('status');
const myIdEl = document.getElementById('myId');
const uiEl = document.getElementById('ui');

const gameContainer = document.getElementById('game-container');

const style = document.createElement('style');
style.textContent = `
    .panel {
        position: absolute;
        background: rgba(0, 0, 0, 0.8);
        border: 2px solid #555;
        padding: 10px;
        color: white;
        font-family: monospace;
        pointer-events: auto;
    }
    .inventory-panel {
        bottom: 10px;
        right: 10px;
        width: 200px;
        height: 150px;
    }
    .equipment-panel {
        bottom: 10px;
        left: 10px;
        width: 200px;
        height: 150px;
    }
    .slot-grid {
        display: grid;
        grid-template-columns: repeat(5, 1fr);
        gap: 5px;
        margin-top: 5px;
    }
    .slot {
        width: 30px;
        height: 30px;
        border: 1px solid #777;
        background: #222;
        display: flex;
        justify-content: center;
        align-items: center;
        cursor: pointer;
        font-size: 10px;
        text-align: center;
        overflow: hidden;
    }
    .slot:hover {
        border-color: #fff;
    }
    .slot.filled {
        background: #444;
    }
    .item-gold { color: gold; }
    .item-weapon { color: cyan; }
    .item-armor { color: violet; }
`;
document.head.appendChild(style);

const inventoryEl = document.createElement('div');
inventoryEl.className = 'panel inventory-panel';
inventoryEl.innerHTML = '<div>Inventory</div><div class="slot-grid" id="inv-grid"></div>';
gameContainer.appendChild(inventoryEl);

const equipmentEl = document.createElement('div');
equipmentEl.className = 'panel equipment-panel';
equipmentEl.innerHTML = '<div>Equipment (Click to Unequip)</div><div class="slot-grid" id="equip-grid"></div>';
gameContainer.appendChild(equipmentEl);

let inventory = [];
let equipment = {};

function renderInventory() {
    const grid = document.getElementById('inv-grid');
    grid.innerHTML = '';
    
    for (let i = 0; i < 20; i++) {
        const slot = document.createElement('div');
        slot.className = 'slot';
        
        if (i < inventory.length) {
            const item = inventory[i];
            slot.classList.add('filled');
            slot.textContent = item.Name ? item.Name.substring(0, 4) : '???';
            slot.title = `${item.Name} (ATK:${item.Attack} DEF:${item.Defense})`;
            
            if (item.Type === 1) slot.style.color = 'cyan';
            if (item.Type === 2) slot.style.color = 'violet';

            slot.onclick = () => {
                let targetSlot = -1;
                for (let s = 0; s < 5; s++) {
                    if (!equipment[s]) {
                        targetSlot = s;
                        break;
                    }
                }
                if (targetSlot === -1) targetSlot = 0;

                console.log(`Equipping item ${item.ID} to slot ${targetSlot}`);
                ws.send(JSON.stringify({
                    type: "EQUIP",
                    item_id: item.ID,
                    slot: targetSlot
                }));
            };
        }
        
        grid.appendChild(slot);
    }
}

function renderEquipment() {
    const grid = document.getElementById('equip-grid');
    grid.innerHTML = '';
    
    for (let i = 0; i < 5; i++) {
        const slot = document.createElement('div');
        slot.className = 'slot';
        
        const item = equipment[i];
        if (item) {
            slot.classList.add('filled');
            slot.textContent = item.Name ? item.Name.substring(0, 4) : 'Item';
            slot.title = `${item.Name}`;
            
             if (item.Type === 1) slot.style.color = 'cyan';
             if (item.Type === 2) slot.style.color = 'violet';

            slot.onclick = () => {
                console.log(`Unequipping slot ${i}`);
                ws.send(JSON.stringify({
                    type: "UNEQUIP",
                    slot: i
                }));
            };
        } else {
            slot.textContent = i + 1;
            slot.style.color = '#555';
        }
        
        grid.appendChild(slot);
    }
}


// Stats UI
const statsEl = document.createElement('div');
statsEl.innerHTML = 'HP: -/- | ATK: - | DEF: - | SPD: - | GOLD: 0';
uiEl.appendChild(statsEl);

const mapNameEl = document.createElement('div');
mapNameEl.innerHTML = 'Map: town';
uiEl.appendChild(mapNameEl);

const players = new Map();
const items = new Map();
const monsters = new Map();
const projectiles = new Map();
let portals = [];

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
                        projectiles.set(p.id, { x: p.x, y: p.y, type: p.type });
                    } else {
                        const existing = projectiles.get(p.id);
                        existing.x = p.x;
                        existing.y = p.y;
                        existing.type = p.type;
                    }
                });
            }
            for (const [id] of projectiles) { if (!seenProjs.has(id)) projectiles.delete(id); }
            break;

        case 'INVENTORY':
            inventory = msg.items || [];
            renderInventory();
            break;

        case 'EQUIPMENT':
            equipment = msg.items || {};
            renderEquipment();
            break;

        case 'STATS':
            // Update stats
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

        case 'ITEM_SPAWN':
            items.set(msg.id, { x: msg.x, y: msg.y, type: msg.item_type });
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

        case 'MAP_SWITCH':
            items.clear();
            monsters.clear();
            projectiles.clear();
            portals = msg.portals || [];
            
            players.forEach((_, id) => {
                if (id !== myId) players.delete(id);
            });
            
            const me = players.get(myId);
            if (me) {
                me.x = msg.x;
                me.y = msg.y;
            }
            
            mapNameEl.textContent = `Map: ${msg.map}`;
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

    items.forEach((item) => {
        if (item.type === 0) ctx.fillStyle = '#ffd700';
        else if (item.type === 1) ctx.fillStyle = 'cyan';
        else if (item.type === 2) ctx.fillStyle = 'violet';
        else ctx.fillStyle = '#fff';

        ctx.fillRect(item.x - 5, item.y - 5, 10, 10);
    });

    ctx.fillStyle = '#800080';
    portals.forEach((p) => {
        ctx.beginPath();
        ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2);
        ctx.fill();
        
        ctx.fillStyle = '#fff';
        ctx.font = '12px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(`To ${p.target}`, p.x, p.y + 5);
        ctx.fillStyle = '#800080';
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

    projectiles.forEach((p) => {
        if (p.type === 1) ctx.fillStyle = 'orange';
        else if (p.type === 2) ctx.fillStyle = 'aqua';
        else ctx.fillStyle = '#ff0';

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
