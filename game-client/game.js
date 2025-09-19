// Nakama Tic-Tac-Toe Web Client
class NakamaGame {
    constructor() {
        this.client = null;
        this.socket = null;
        this.session = null;
        this.matchId = null;
        this.mySymbol = "";
        this.currentTurn = "";
        this.playerCount = 0;
        this.gameBoard = Array(9).fill("");
        this.isMyTurn = false;
        this.gameEnded = false;
        this.playerList = [];

        this.initializeUI();
    }

    initializeUI() {
        // Get DOM elements
        this.elements = {
            connectBtn: document.getElementById('connect-btn'),
            disconnectBtn: document.getElementById('disconnect-btn'),
            username: document.getElementById('username'),
            gameMode: document.getElementById('game-mode'),
            serverUrl: document.getElementById('server-url'),
            deviceId: document.getElementById('device-id'),
            statusIndicator: document.getElementById('status-indicator'),
            connectionStatus: document.getElementById('connection-status'),
            gamePanel: document.getElementById('game-panel'),
            gameBoard: document.getElementById('game-board'),
            mySymbol: document.getElementById('my-symbol'),
            currentTurn: document.getElementById('current-turn'),
            playerCount: document.getElementById('player-count'),
            currentGameMode: document.getElementById('current-game-mode'),
            gameState: document.getElementById('game-state'),
            timerDisplay: document.getElementById('timer-display'),
            timeRemaining: document.getElementById('time-remaining'),
            timerProgress: document.getElementById('timer-progress'),
            messages: document.getElementById('messages')
        };
        this.elements.newGameBtn = document.getElementById('new-game-btn');
        this.elements.newGameBtn.addEventListener('click', () => {
            this.elements.newGameBtn.classList.add('hidden');
            this.hideGameUI();
            this.startMatchmaking();
        });

        this.leaderboardPanel = document.getElementById('leaderboard-panel');
        this.leaderboardTable = document.getElementById('leaderboard-table');

        // Add event listeners
        this.elements.connectBtn.addEventListener('click', () => this.connect());
        this.elements.disconnectBtn.addEventListener('click', () => this.disconnect());

        // Add click listeners to game board cells
        document.querySelectorAll('.cell').forEach(cell => {
            cell.addEventListener('click', (e) => this.handleCellClick(e));
        });
    }

    async connect() {
        try {
            // Check if nakama is available in various possible locations
            let Nakama = null;
            if (window.nakama && window.nakama.Client) {
                Nakama = window.nakama;
            } else if (window.NakamaJs && window.NakamaJs.Client) {
                Nakama = window.NakamaJs;
            } else if (window.Client) {
                // Sometimes Client is exposed directly
                Nakama = { Client: window.Client };
            } else {
                console.error('❌ Nakama library not found. Available:', Object.keys(window).filter(k => k.toLowerCase().includes('nakama')));
                throw new Error('Nakama library not loaded. Please refresh the page and check your internet connection.');
            }

            this.updateStatus('connecting', 'Connecting...');
            this.addMessage('info', '🔌 Connecting to Nakama server...');

            const username = this.elements.username.value.trim();
            const gameMode = this.elements.gameMode.value;
            const serverUrl = this.elements.serverUrl.value.trim();
            const deviceId = this.elements.deviceId.value.trim();

            if (!username || !serverUrl || !deviceId) {
                throw new Error('Please enter username, server URL and device ID');
            }

            // Store selected game mode
            this.selectedGameMode = gameMode;
            this.addMessage('info', `🎮 Selected ${gameMode} mode`);

            // Parse server URL
            const [host, port] = serverUrl.split(':');
            
            // Initialize Nakama client
            this.client = new Nakama.Client("defaultkey", host, port ? parseInt(port) : 7350, false);

            // Authenticate
            this.session = await this.client.authenticateDevice(deviceId, true, username);
            this.addMessage('success', `✅ Authenticated as: ${username} (${this.session.user_id})`);

            // Connect socket
            this.socket = this.client.createSocket();
            await this.socket.connect(this.session);

            this.updateStatus('connected', 'Connected');
            this.addMessage('success', '🎮 Socket connected! Looking for match...');

            this.setupSocketHandlers();
            await this.startMatchmaking();

            // Fetch leaderboard after connecting and run this every ten seconds
            this.fetchLeaderboard();
            setInterval(() => this.fetchLeaderboard(), 10000);

        } catch (error) {
            this.updateStatus('disconnected', 'Connection Failed');
            this.addMessage('error', `❌ Connection failed: ${error.message}`);
            console.error('Connection error:', error);
        }
    }

    setupSocketHandlers() {
        // Handle match data
        this.socket.onmatchdata = (matchData) => {
            const opCode = matchData.op_code;
            const data = JSON.parse(new TextDecoder().decode(matchData.data));
            
            switch (opCode) {
                case 1: // Welcome message
                    this.addMessage('game', `💌 ${data.message}`);
                    if (data.game_mode) {
                        this.currentGameMode = data.game_mode;
                        this.elements.currentGameMode.textContent = data.game_mode;
                        this.showTimerIfNeeded(data.game_mode);
                    }
                    break;
                    
                case 2: // Player joined
                    this.handlePlayerJoined(data);
                    this.playerList.push(data.user_id);
                    break;
                    
                case 3: // Player left
                    this.addMessage('game', `👋 ${data.message}`);
                    this.playerList = this.playerList.filter(id => id !== data.user_id);
                    break;
                    
                case 4: // Game update
                    this.handleGameUpdate(data);
                    break;
                    
                case 5: // Game over
                    this.handleGameOver(data);
                    break;

                case 6: // Player error
                    this.addMessage('error', `❌ Error: ${data.error}`);
                    break;

                case 7: // Game draw
                    this.handleGameDraw(data);
                    break;

                case 8: // Timeout win
                    this.handleTimeoutWin(data);
                    break;

                case 9: // Timer update
                    this.handleTimerUpdate(data);
                    break;
                    
                default:
                    this.addMessage('info', `🔍 Unknown message: ${JSON.stringify(data)}`);
            }
        };

        // Handle matchmaker
        this.socket.onmatchmakermatched = async (matched) => {
            try {
                this.addMessage('success', `🎯 Match found! Joining match...`);
                const match = await this.socket.joinMatch(matched.match_id, matched.token);
                this.matchId = match.match_id;
                this.addMessage('success', `🏆 Joined match: ${this.matchId}`);
                this.updateStatus('connected', 'In Match');
                // Show game UI
                this.showGameUI();
                
            } catch (error) {
                this.addMessage('error', `❌ Failed to join match: ${error.message}`);
            }
        };

        // Handle disconnection
        this.socket.ondisconnect = () => {
            this.updateStatus('disconnected', 'Disconnected');
            this.hideGameUI();
            this.addMessage('error', '🔌 Disconnected from server');
        };
    }

    async startMatchmaking() {
        try {
            this.updateStatus('matching', 'Finding match...');
            
            // Use wildcard query and rely on backend to match by properties
            const query = "properties.mode:" + this.selectedGameMode + "";
            const minCount = 2;
            const maxCount = 2;
            
            // Add properties to include game mode in matchmaking
            const properties = {
                "mode": this.selectedGameMode
            };
            
            await this.socket.addMatchmaker(query, minCount, maxCount, properties, null);
            this.addMessage('info', `🔍 Looking for ${this.selectedGameMode} mode opponents (2 players needed)...`);
            
        } catch (error) {
            this.addMessage('error', `❌ Matchmaking failed: ${error.message}`);
        }
    }

    handlePlayerJoined(data) {
        this.addMessage('game', `👤 ${data.username || data.user_id} joined the game!`);
        
        if (data.user_id === this.session.user_id) {
            this.mySymbol = data.symbol;
            this.elements.mySymbol.textContent = this.mySymbol;
            this.addMessage('success', `🎯 Your symbol is: ${this.mySymbol}`);
        }
        
        this.playerCount = data.total_players;
        this.elements.playerCount.textContent = this.playerCount;
        
        if (data.current_turn) {
            this.currentTurn = data.current_turn;
            this.updateTurnDisplay();
        }
        
        if (data.board_state) {
            this.updateBoard(data.board_state);
        }
        
        if (this.playerCount === 2) {
            this.elements.gameState.textContent = "Game Started!";
            this.addMessage('success', '🚀 Game started! Make your move!');
        }
    }

    handleGameUpdate(data) {
        if (data.board_state) {
            this.updateBoard(data.board_state);
        }
        
        if (data.current_turn) {
            this.currentTurn = data.current_turn;
            this.updateTurnDisplay();
        }

        if (data.game_mode) {
            this.currentGameMode = data.game_mode;
            this.elements.currentGameMode.textContent = data.game_mode;
            this.showTimerIfNeeded(data.game_mode);
        }
        
        // Update timer if in timed mode
        if (data.time_remaining !== undefined && this.currentGameMode === 'timed') {
            this.addMessage('server', `Server time remaining: ${data.time_remaining}s`);
            this.updateTimer(data.time_remaining);
        }
        this.addMessage('game', '🔄 Board updated');
    }

    handleGameOver(data) {
    this.elements.newGameBtn.classList.remove('hidden');
        this.gameEnded = true;
        this.addMessage('game', `🏆 ${data.message}`);
        this.elements.gameState.textContent = `Game Over - ${data.message}`;
        
        if (data.board_state) {
            this.updateBoard(data.board_state);
        }
        
        // Hide timer if it was showing
        this.hideTimer();
        
        // Disable all cells
        document.querySelectorAll('.cell').forEach(cell => {
            cell.disabled = true;
        });
    }

    handleGameDraw(data) {
    this.elements.newGameBtn.classList.remove('hidden');
        this.gameEnded = true;
        this.addMessage('game', `🤝 ${data.message}`);
        this.elements.gameState.textContent = `Game Draw - ${data.message}`;
        
        if (data.board_state) {
            this.updateBoard(data.board_state);
        }
        
        // Hide timer if it was showing
        this.hideTimer();
        
        // Disable all cells
        document.querySelectorAll('.cell').forEach(cell => {
            cell.disabled = true;
        });
    }

    handleTimeoutWin(data) {
    this.elements.newGameBtn.classList.remove('hidden');
            this.elements.newGameBtn.classList.add('hidden');
            this.elements.newGameBtn.classList.add('hidden');
        this.gameEnded = true;
        this.addMessage('error', `⏰ ${data.message}`);
        this.elements.gameState.textContent = `Timeout - ${data.message}`;
        
        if (data.board_state) {
            this.updateBoard(data.board_state);
        }
        
        // Hide timer
        this.hideTimer();
        
        // Disable all cells
        document.querySelectorAll('.cell').forEach(cell => {
            cell.disabled = true;
        });
    }

    handleTimerUpdate(data) {
        if (this.currentGameMode === 'timed') {
            // this.addMessage('server', `Server time remaining: ${data.time_remaining}s`);
            this.updateTimer(data.time_remaining);
        }
    }

    showTimerIfNeeded(gameMode) {
        if (gameMode === 'timed') {
            this.elements.timerDisplay.classList.remove('hidden');
        } else {
            this.elements.timerDisplay.classList.add('hidden');
        }
    }

    updateTimer(timeRemaining) {
        this.elements.timeRemaining.textContent = Math.max(0, timeRemaining);
        // Debug: show timer update
        this.addMessage('info', `⏰ Timer update: ${Math.max(0, timeRemaining)}s left`);
        // Update progress bar (assuming 30 second max)
        const percentage = Math.max(0, (timeRemaining / 30) * 100);
        this.elements.timerProgress.style.width = `${percentage}%`;
        // Add warning style if time is running low
        if (timeRemaining <= 10) {
            this.elements.timerProgress.classList.add('timer-warning');
        } else {
            this.elements.timerProgress.classList.remove('timer-warning');
        }
    }

    hideTimer() {
        this.elements.timerDisplay.classList.add('hidden');
    }

    updateBoard(boardState) {
        this.gameBoard = boardState;
        const cells = document.querySelectorAll('.cell');
        
        cells.forEach((cell, index) => {
            const symbol = boardState[index];
            cell.textContent = symbol;
            cell.className = `cell ${symbol.toLowerCase()}`;
            cell.disabled = symbol !== "" || this.gameEnded || !this.isMyTurn;
        });
    }

    updateTurnDisplay() {
        this.isMyTurn = this.currentTurn === this.session.user_id;
        
        if (this.isMyTurn) {
            this.elements.currentTurn.textContent = "Your turn!";
            this.elements.gameState.textContent = "Your turn - click a cell!";
            this.addMessage('success', '🎯 It\'s your turn!');
        } else {
            this.elements.currentTurn.textContent = "Opponent's turn";
            this.elements.gameState.textContent = "Waiting for opponent...";
        }
        
        // Update cell disabled state
        document.querySelectorAll('.cell').forEach((cell, index) => {
            const isEmpty = this.gameBoard[index] === "";
            cell.disabled = !isEmpty || !this.isMyTurn || this.gameEnded;
        });
    }

    async handleCellClick(event) {
        if (!this.isMyTurn || this.gameEnded || !this.socket || !this.matchId) {
            return;
        }
        
        const row = parseInt(event.target.dataset.row);
        const col = parseInt(event.target.dataset.col);
        const index = row * 3 + col;
        
        if (this.gameBoard[index] !== "") {
            this.addMessage('error', '❌ Cell already occupied!');
            return;
        }
        
        try {
            const moveData = { row, col };
            await this.socket.sendMatchState(this.matchId, 1, JSON.stringify(moveData));
            this.addMessage('info', `📤 Move sent: ${this.mySymbol} to (${row}, ${col})`);
            
        } catch (error) {
            this.addMessage('error', `❌ Failed to send move: ${error.message}`);
        }
    }

    showGameUI() {
        this.elements.gamePanel.classList.remove('hidden');
        this.elements.gameBoard.classList.remove('hidden');
    }

    hideGameUI() {
        this.elements.gamePanel.classList.add('hidden');
        this.elements.gameBoard.classList.add('hidden');
        this.resetGame();
    }

    resetGame() {
        this.matchId = null;
        this.mySymbol = "";
        this.currentTurn = "";
        this.playerCount = 0;
        this.gameBoard = Array(9).fill("");
        this.isMyTurn = false;
        this.gameEnded = false;
        
        // Reset UI
        this.elements.mySymbol.textContent = "-";
        this.elements.currentTurn.textContent = "-";
        this.elements.playerCount.textContent = "0";
        this.elements.gameState.textContent = "Waiting for players...";
        
        // Reset board
        document.querySelectorAll('.cell').forEach(cell => {
            cell.textContent = "";
            cell.className = "cell";
            cell.disabled = false;
        });
    }

    async disconnect() {
        try {
            if (this.socket) {
                this.socket.disconnect();
            }
            
            this.updateStatus('disconnected', 'Disconnected');
            this.hideGameUI();
            this.addMessage('info', '👋 Disconnected from server');
            
        } catch (error) {
            this.addMessage('error', `❌ Disconnect error: ${error.message}`);
        }
    }

    updateStatus(status, text) {
        this.elements.connectionStatus.textContent = text;
        this.elements.statusIndicator.className = `status-indicator status-${status}`;
        
        // Toggle buttons
        if (status === 'connected' || status === 'matching') {
            this.elements.connectBtn.classList.add('hidden');
            this.elements.disconnectBtn.classList.remove('hidden');
            this.elements.serverUrl.disabled = true;
            this.elements.deviceId.disabled = true;
            this.elements.username.disabled = true;
            this.elements.gameMode.disabled = true;
        } else {
            this.elements.connectBtn.classList.remove('hidden');
            this.elements.disconnectBtn.classList.add('hidden');
            this.elements.serverUrl.disabled = false;
            this.elements.username.disabled = true;
            this.elements.gameMode.disabled = false;
        }
    }

    addMessage(type, message) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;
        messageDiv.textContent = `${new Date().toLocaleTimeString()} - ${message}`;
        
        this.elements.messages.appendChild(messageDiv);
        this.elements.messages.scrollTop = this.elements.messages.scrollHeight;
        
        // Keep only last 50 messages
        while (this.elements.messages.children.length > 50) {
            this.elements.messages.removeChild(this.elements.messages.firstChild);
        }
    }

    async fetchLeaderboard() {
        if (!this.client || !this.session) return;
        try {
            // Call the custom RPC to get top N players
            const n = 10;
            const rpcResponse = await this.client.rpc(this.session, 'GetTopPlayers', JSON.stringify({ n }));
            const players = rpcResponse.payload || [];
            this.renderLeaderboard(players);
        } catch (error) {
            this.renderEmptyLeaderboard();
        }
    }

    renderLeaderboard(players) {
        if (!this.leaderboardTable) return;
        const tbody = this.leaderboardTable.querySelector('tbody');
        tbody.innerHTML = '';
        players.forEach((player, idx) => {
            const tr = document.createElement('tr');
            tr.innerHTML = `<td>${idx + 1}</td><td>${player.username || player.owner_id}</td><td>${player.score}</td>`;
            tbody.appendChild(tr);
        });
    }

    renderEmptyLeaderboard() {
        if (!this.leaderboardTable) return;
        const tbody = this.leaderboardTable.querySelector('tbody');
        tbody.innerHTML = '<tr><td colspan="3" style="text-align:center;">No scores yet - play some games!</td></tr>';
    }
}

// Initialize the game when the page loads
document.addEventListener('DOMContentLoaded', () => {
    
    // Wait a bit for Nakama library to be fully loaded
    setTimeout(() => {
        try {
            window.game = new NakamaGame();
        } catch (error) {
            console.error('❌ Failed to initialize game:', error);
        }
    }, 100);
});