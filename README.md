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

## Play Online
Try it out here: [Nakama Tic-Tac-Toe Web UI](http://3.109.202.60)
- Just fill in your username and a unique device ID
- Click "Connect" and wait for matchmaking (need 2 players)

## 🚀 Quick Start

### 1. Start Nakama Server
```bash
cd game-server
docker-compose up --build
```

### 2. Start Web Server
```bash
cd game-client
node server.js
```

### 3. Open in Browser
Go to: http://localhost:3000

### 4. Play!
1. Enter server details (default: `127.0.0.1:7350`)
2. Enter Username, this will be attached to the device ID
3. Choose the game mode, timed or classic.
4. Enter a device ID (any unique string)
5. Click "Connect"
6. Wait for matchmaking (need 2 players)
7. Click cells to make moves!

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
game-client/
├── index.html      # Main UI layout and styles
├── game.js         # Game logic and Nakama integration  
├── server.js       # Simple HTTP server
├── nakama.js       # Original terminal client (still works)
```

```
game-server/
├── docker-compose.yml  # Nakama server setup
├── main.go            # Custom server logic (if any)
├── match.go           # Match handler logic and game logic
├── go.mod             # Go module file
├── go.sum             # Go dependencies

```

## 🔧 Technical Details

### FrontEnd
- **Frontend**: Vanilla HTML/CSS/JavaScript
- **Nakama Client**: CDN version for browser compatibility
- **Server**: Nginx HTTP server
- **Styling**: CSS Grid for game board, CSS animations
- **Real-time**: WebSocket connection to Nakama

### Backend
- **Nakama**: Open-source multiplayer server
- **Match Handler**: Custom Go code for tic-tac-toe logic
- **Database**: PostgreSQL for persistence
- **Containerization**: Docker and Docker Compose for easy setup 
- **Ports**: Nakama (7350, 7351), PostgreSQL (5432)
- **Data Persistence**: Docker volumes for Nakama and PostgreSQL data


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
- Add spectator mode
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