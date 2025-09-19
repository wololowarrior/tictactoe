# 🎮 Nakama Tic-Tac-Toe Web UI

A beautiful, interactive web interface for playing real-time tic-tac-toe using Nakama multiplayer server.

## ✨ Features

- 🌐 **Web-based UI** - No terminal commands needed!
- 🎯 **Visual Game Board** - Click cells to make moves
- ⚡ **Real-time Multiplayer** - Instant updates across players
- 🎨 **Modern Design** - Beautiful gradient UI with animations
- 📱 **Responsive** - Works on desktop and mobile
- 🔌 **Connection Management** - Easy connect/disconnect controls
- 📊 **Game Status** - Turn indicators and player info
- 💬 **Live Messages** - Real-time game events and status

## 🚀 Quick Start

### 1. Start Nakama Server
```bash
cd /Users/harshil.gupta/go/src/lila
docker-compose up --build
```

### 2. Start Web Server
```bash
cd /Users/harshil.gupta/go/src/lila_client
node server.js
```

### 3. Open in Browser
Go to: http://localhost:3000

### 4. Play!
1. Enter server details (default: `127.0.0.1:7350`)
2. Enter a device ID (any unique string)
3. Click "Connect"
4. Wait for matchmaking (need 2 players)
5. Click cells to make moves!

## 🎯 How to Play

- **Goal**: Get 3 of your symbols (X or O) in a row
- **Turns**: Players alternate turns
- **Winning**: First to get 3 in a row (horizontal, vertical, or diagonal) wins!
- **Visual Cues**: 
  - Your turn = cells are clickable
  - Opponent's turn = cells are disabled
  - X symbols appear in red
  - O symbols appear in blue

## 📁 File Structure

```
lila_client/
├── index.html      # Main UI layout and styles
├── game.js         # Game logic and Nakama integration  
├── server.js       # Simple HTTP server
├── nakama.js       # Original terminal client (still works)
└── README.md       # This file
```

## 🔧 Technical Details

- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Nakama Client**: CDN version for browser compatibility
- **Server**: Simple Node.js HTTP server
- **Styling**: CSS Grid for game board, CSS animations
- **Real-time**: WebSocket connection to Nakama

## 🐛 Troubleshooting

### Can't Connect
- Make sure Nakama server is running on `localhost:7350`
- Check browser console for errors
- Verify device ID is unique

### Game Doesn't Start
- Need exactly 2 players to start a match
- Open multiple browser tabs/windows to test locally
- Check Nakama server logs for errors

### Moves Not Working
- Make sure it's your turn (check status panel)
- Cells must be empty to click
- Check connection status

## 🚀 Next Steps

- Add player names/avatars
- Add game history/scores
- Add spectator mode
- Add different game modes
- Add sound effects
- Add mobile optimizations

## 🤝 Multiplayer Testing

To test locally:
1. Open http://localhost:3000 in multiple browser tabs
2. Use different device IDs for each tab
3. Connect all tabs
4. Matchmaking will pair them up!

---

**Enjoy playing! 🎉**