package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

// GameMode represents different game modes
type GameMode string

const (
	GameModeClassic GameMode = "classic"
	GameModeTimed   GameMode = "timed"
)

// MatchState represents the persistent state of the match
type MatchState struct {
	Players       []runtime.Presence `json:"players"`
	PlayerActions map[string][]byte  `json:"player_actions"`
	TicTacToe     [9]string          `json:"tictactoe"`
	PlayerSymbols map[string]string  `json:"player_symbols"`
	CurrentTurn   string             `json:"current_turn"`
	GameStarted   bool               `json:"game_started"`
	GameEnded     bool               `json:"game_ended"`
	Winner        string             `json:"winner,omitempty"`

	// Game mode specific fields
	GameMode         GameMode `json:"game_mode"`
	TurnTimeLimit    int64    `json:"turn_time_limit,omitempty"`    // seconds per turn
	CurrentTurnStart int64    `json:"current_turn_start,omitempty"` // timestamp when current turn started
	TimeRemaining    int64    `json:"time_remaining,omitempty"`     // seconds remaining for current turn
}

// Match represents our custom match implementation (now just configuration)
type Match struct {
	matchLabel  string
	tickRate    int
	gameMode    GameMode
	marshaler   *protojson.MarshalOptions
	unmarshaler *protojson.UnmarshalOptions
}

func NewMatch() *Match {
	return &Match{
		matchLabel: "lobby",
		tickRate:   10,
		gameMode:   GameModeClassic,
		marshaler: &protojson.MarshalOptions{
			UseEnumNumbers: true,
		},
		unmarshaler: &protojson.UnmarshalOptions{
			DiscardUnknown: false,
		},
	}
}

// NewMatchWithMode creates a match with specific game mode
func NewMatchWithMode(mode GameMode) *Match {
	match := NewMatch()
	match.gameMode = mode
	match.matchLabel = fmt.Sprintf("lobby_%s", mode)
	return match
}

// newMatchState creates a new initial match state
func newMatchState(gameMode GameMode) *MatchState {
	state := &MatchState{
		Players:       []runtime.Presence{},
		PlayerActions: make(map[string][]byte),
		TicTacToe:     [9]string{"", "", "", "", "", "", "", "", ""},
		PlayerSymbols: make(map[string]string),
		CurrentTurn:   "",
		GameStarted:   false,
		GameEnded:     false,
		Winner:        "",
		GameMode:      gameMode,
	}

	// Set timed mode specific settings
	if gameMode == GameModeTimed {
		state.TurnTimeLimit = 30 // 30 seconds per turn
		state.CurrentTurnStart = 0
		state.TimeRemaining = 30
	}

	return state
}

// getMatchState safely extracts MatchState from interface{}
func getMatchState(state interface{}) *MatchState {
	if state == nil {
		return newMatchState(GameModeClassic)
	}

	if matchState, ok := state.(*MatchState); ok {
		return matchState
	}

	return newMatchState(GameModeClassic)
}

var gameWin = [8][3]int{
	{0, 1, 2}, // Rows
	{3, 4, 5},
	{6, 7, 8},
	{0, 3, 6}, // Columns
	{1, 4, 7},
	{2, 5, 8},
	{0, 4, 8}, // Diagonals
	{2, 4, 6},
}

// MatchInit initializes the match
// MatchInit initializes the match
func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	logger.Info("=== MATCH INITIALIZED ===")
	logger.Info("Match created with tick rate: %d, label: %s", m.tickRate, m.matchLabel)
	logger.Info("Match creation params: %v", params)

	// Check if game mode is specified in params
	gameMode := m.gameMode
	if modeParam, exists := params["game_mode"]; exists {
		if modeStr, ok := modeParam.(string); ok {
			gameMode = GameMode(modeStr)
			logger.Info("Game mode set from params: %s", gameMode)
		}
	}

	// Return initial match state with the specified game mode
	initialState := newMatchState(gameMode)
	logger.Info("Match initialized with mode: %s", initialState.GameMode)
	return initialState, m.tickRate, m.matchLabel
}

// MatchJoinAttempt is called when a user attempts to join the match
func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	logger.Info("=== MATCH JOIN ATTEMPT === Player %s attempting to join match", presence.GetUserId())
	return state, true, ""
}

// MatchJoin is called when a user successfully joins the match
func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	logger.Info("=== MATCH JOIN === Processing %d users joining", len(presences))

	// Get current state
	matchState := getMatchState(state)

	// Add all new players to our list first
	for _, presence := range presences {
		// Prevent duplicate entries
		for _, p := range matchState.Players {
			if p.GetUserId() == presence.GetUserId() {
				logger.Info("Player %s already in match, skipping", presence.GetUserId())
				continue
			}
		}
		matchState.Players = append(matchState.Players, presence)
		logger.Info("=== PLAYER JOINED === Player %s joined match (total players: %d)", presence.GetUserId(), len(matchState.Players))
	}

	// Assign symbols to players if not already assigned
	symbols := []string{"X", "O"}
	for i, player := range matchState.Players {
		if _, exists := matchState.PlayerSymbols[player.GetUserId()]; !exists {
			matchState.PlayerSymbols[player.GetUserId()] = symbols[i%2]
			logger.Info("Assigned symbol %s to player %s", symbols[i%2], player.GetUserId())
			// Set the first player to join as the current turn
			if matchState.CurrentTurn == "" {
				matchState.CurrentTurn = player.GetUserId()
				logger.Info("It's now player %s's turn", matchState.CurrentTurn)

				// Start timer for timed mode
				if matchState.GameMode == GameModeTimed {
					matchState.CurrentTurnStart = time.Now().Unix()
					matchState.TimeRemaining = matchState.TurnTimeLimit
					logger.Info("Started timer for timed mode: %d seconds", matchState.TurnTimeLimit)
				}
			}
		}
	}

	// Send messages to players
	for _, presence := range presences {
		// Welcome message
		messageData := map[string]interface{}{
			"message":      fmt.Sprintf("Welcome to %s mode!", matchState.GameMode),
			"player_count": len(matchState.Players),
			"game_mode":    matchState.GameMode,
		}

		// Add timed mode specific info
		if matchState.GameMode == GameModeTimed {
			messageData["turn_time_limit"] = matchState.TurnTimeLimit
		}

		messageBytes, _ := json.Marshal(messageData)
		dispatcher.BroadcastMessage(1, messageBytes, []runtime.Presence{presence}, nil, true)

		// Game state announcement
		announceData := map[string]interface{}{
			"message":       "New player joined!",
			"user_id":       presence.GetUserId(),
			"username":      presence.GetUsername(),
			"total_players": len(matchState.Players),
			"current_turn":  matchState.CurrentTurn,
			"board_state":   matchState.TicTacToe,
			"symbol":        matchState.PlayerSymbols[presence.GetUserId()],
			"game_mode":     matchState.GameMode,
		}

		// Add timed mode specific data
		if matchState.GameMode == GameModeTimed {
			announceData["turn_time_limit"] = matchState.TurnTimeLimit
			announceData["time_remaining"] = matchState.TimeRemaining
		}

		announceBytes, _ := json.Marshal(announceData)
		dispatcher.BroadcastMessage(2, announceBytes, matchState.Players, nil, true)
	}

	return matchState
}

// MatchLeave is called when a user leaves the match
func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	matchState := getMatchState(state)

	for _, presence := range presences {
		// Remove the player from our list
		for i, p := range matchState.Players {
			if p.GetUserId() == presence.GetUserId() {
				matchState.Players = append(matchState.Players[:i], matchState.Players[i+1:]...)
				break
			}
		}

		// Announce player departure
		announceData := map[string]interface{}{
			"message": "Player left the match",
			"user_id": presence.GetUserId(),
		}
		announceBytes, _ := json.Marshal(announceData)

		// Send a message to all remaining players
		dispatcher.BroadcastMessage(3, announceBytes, nil, nil, true)

		logger.Info("Player %s left match", presence.GetUserId())
	}
	return matchState
}

// MatchLoop is called on every tick
func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	matchState := getMatchState(state)

	// Handle timer for timed mode
	if matchState.GameMode == GameModeTimed && matchState.CurrentTurnStart > 0 && !matchState.GameEnded {
		currentTime := time.Now().Unix()
		elapsed := currentTime - matchState.CurrentTurnStart
		matchState.TimeRemaining = matchState.TurnTimeLimit - elapsed
		// Check if time is up
		if matchState.TimeRemaining <= 0 {
			logger.Info("Time's up for player %s", matchState.CurrentTurn)

			// Current player loses due to timeout
			var winner string
			var winnerSymbol string
			for _, p := range matchState.Players {
				if p.GetUserId() != matchState.CurrentTurn {
					winner = p.GetUserId()
					winnerSymbol = matchState.PlayerSymbols[winner]
					break
				}
			}

			if winner != "" {
				timeoutData := map[string]interface{}{
					"message":     fmt.Sprintf("Time's up! %s wins by timeout!", winnerSymbol),
					"winner_id":   winner,
					"board_state": matchState.TicTacToe,
					"game_mode":   matchState.GameMode,
					"timeout":     true,
				}
				timeoutBytes, _ := json.Marshal(timeoutData)
				dispatcher.BroadcastMessage(8, timeoutBytes, nil, nil, true)

				matchState.GameEnded = true
				matchState.Winner = winner

				// Write to leaderboard
				m.writeToLeaderboard(ctx, nk, logger, winner, winnerSymbol, matchState)
				return matchState
			}
		}

		// Broadcast time update every few seconds
		if tick%10 == 0 { // Every 10 ticks (assuming 10 ticks per second)
			timeData := map[string]interface{}{
				"time_remaining": matchState.TimeRemaining,
				"current_turn":   matchState.CurrentTurn,
			}
			timeBytes, _ := json.Marshal(timeData)
			dispatcher.BroadcastMessage(9, timeBytes, nil, nil, true)
		}
	}

	// Process any messages from players
	for _, message := range messages {
		// Get player's presence from state
		var presence runtime.Presence
		for _, p := range matchState.Players {
			if p.GetUserId() == message.GetUserId() {
				presence = p
				break
			}
		}

		logger.Info("Received message from user %s: %v", message.GetUserId(), string(message.GetData()))
		matchState.PlayerActions[message.GetUserId()] = message.GetData()

		// Parse the action
		var action struct {
			Row int `json:"row"`
			Col int `json:"col"`
		}
		if err := json.Unmarshal(message.GetData(), &action); err != nil {
			logger.Error("Invalid action data from user %s: %v", message.GetUserId(), err)
			playerError(dispatcher, presence, "Invalid action data")
			continue
		}

		// Validate move
		if action.Row < 0 || action.Row > 2 || action.Col < 0 || action.Col > 2 {
			logger.Error("Out of bounds move from user %s: %v", message.GetUserId(), action)
			playerError(dispatcher, presence, "Out of bounds move")
			continue
		}
		if matchState.TicTacToe[action.Row*3+action.Col] != "" {
			logger.Error("Cell already occupied by user %s: %v", message.GetUserId(), action)
			playerError(dispatcher, presence, "Cell already occupied")
			continue
		}
		if matchState.CurrentTurn != "" && matchState.CurrentTurn != message.GetUserId() {
			logger.Error("Not user %s's turn", message.GetUserId())
			playerError(dispatcher, presence, "Not your turn")
			continue
		}

		// Make the move
		symbol := matchState.PlayerSymbols[message.GetUserId()]
		matchState.TicTacToe[action.Row*3+action.Col] = symbol
		logger.Info("Board state: %s, symbol: %s", matchState.TicTacToe, symbol)

		// Check for win
		for _, win := range gameWin {
			if matchState.TicTacToe[win[0]] == symbol && matchState.TicTacToe[win[1]] == symbol && matchState.TicTacToe[win[2]] == symbol {
				// We have a winner
				winData := map[string]interface{}{
					"message":     fmt.Sprintf("We have a winner! %s wins in %s mode!", symbol, matchState.GameMode),
					"winner_id":   message.GetUserId(),
					"board_state": matchState.TicTacToe,
					"game_mode":   matchState.GameMode,
				}
				winBytes, _ := json.Marshal(winData)
				dispatcher.BroadcastMessage(5, winBytes, nil, nil, true)

				matchState.GameEnded = true
				matchState.Winner = message.GetUserId()

				// Write to leaderboard
				m.writeToLeaderboard(ctx, nk, logger, message.GetUserId(), symbol, matchState)
				return matchState
			}
		}

		// Check for draw
		isDraw := true
		for _, cell := range matchState.TicTacToe {
			if cell == "" {
				isDraw = false
				break
			}
		}
		if isDraw {
			drawData := map[string]interface{}{
				"message":     fmt.Sprintf("It's a draw in %s mode!", matchState.GameMode),
				"board_state": matchState.TicTacToe,
				"game_mode":   matchState.GameMode,
			}
			drawBytes, _ := json.Marshal(drawData)
			dispatcher.BroadcastMessage(7, drawBytes, nil, nil, true)

			matchState.GameEnded = true
			logger.Info("Game ended in a draw")
			return matchState
		}

		// Switch turn to the other player
		if len(matchState.Players) == 2 {
			for _, p := range matchState.Players {
				if p.GetUserId() != message.GetUserId() {
					matchState.CurrentTurn = p.GetUserId()
					break
				}
			}

			// Reset timer for timed mode
			if matchState.GameMode == GameModeTimed {
				matchState.CurrentTurnStart = time.Now().Unix()
				matchState.TimeRemaining = matchState.TurnTimeLimit
			}
		}

		// Broadcast game update
		echoData := map[string]interface{}{
			"board_state":  matchState.TicTacToe,
			"current_turn": matchState.CurrentTurn,
			"game_mode":    matchState.GameMode,
		}

		if matchState.GameMode == GameModeTimed {
			echoData["time_remaining"] = matchState.TimeRemaining
		}

		echoBytes, _ := json.Marshal(echoData)
		dispatcher.BroadcastMessage(4, echoBytes, nil, nil, true)
	}

	// Clear actions after processing
	matchState.PlayerActions = make(map[string][]byte)
	return matchState
}

// writeToLeaderboard writes the winner to the leaderboard with mode-specific scoring
func (m *Match) writeToLeaderboard(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger, winnerId, symbol string, matchState *MatchState) {
	score := int64(1) // Default score for classic mode

	// Timed mode gets bonus points
	if matchState.GameMode == GameModeTimed {
		score = 2 // Timed mode is worth more points
	}

	// Get username
	username := ""
	for _, p := range matchState.Players {
		if p.GetUserId() == winnerId {
			username = p.GetUsername()
			break
		}
	}

	metadata := map[string]interface{}{
		"Symbol": symbol,
		"Mode":   matchState.GameMode,
	}

	if matchState.GameMode == GameModeTimed {
		metadata["TimeRemaining"] = matchState.TimeRemaining
	}

	_, err := nk.LeaderboardRecordWrite(ctx, "TicTacToeLeaderboard", winnerId, username, score, 0, metadata, nil)
	if err != nil {
		logger.Error("Failed to write leaderboard record: %v", err)
	} else {
		logger.Info("Wrote leaderboard record for %s: %d points (%s mode)", username, score, matchState.GameMode)
	}
}

func playerError(dispatcher runtime.MatchDispatcher, presence runtime.Presence, message string) {
	errorData := map[string]interface{}{
		"error": message,
	}
	errorBytes, _ := json.Marshal(errorData)
	dispatcher.BroadcastMessage(6, errorBytes, []runtime.Presence{presence}, nil, true)
}

// MatchTerminate is called when the match is terminated
func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	logger.Info("Match terminated")
	return state
}

// MatchSignal is called when the match receives a signal
func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	logger.Info("Match signal received: %v", data)
	return state, data
}
