package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	errInternal = runtime.NewError("internal server error", 13)
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	initializer.RegisterBeforeRt("MatchmakerAdd", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
		req, ok := in.Message.(*rtapi.Envelope_MatchmakerAdd)
		if !ok {
			return nil, errInternal
		}

		logger.Info("=== MATCHMAKER ADD REQUEST ===")
		logger.Info("Original query: %s", req.MatchmakerAdd.Query)
		logger.Info("String properties: %v", req.MatchmakerAdd.StringProperties)
		logger.Info("Numeric properties: %v", req.MatchmakerAdd.NumericProperties)

		return in, nil
	})

	if err := initializer.RegisterMatchmakerMatched(func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, entries []runtime.MatchmakerEntry) (string, error) {
		// Check that all entries have the same game mode
		if len(entries) == 0 {
			return "", runtime.NewError("no matchmaker entries", 3)
		}

		// Get game mode from first entry
		gameMode := "classic"
		logger.Info("Matchmaker entry properties: %v", entries[0].GetProperties())
		if mode, ok := entries[0].GetProperties()["mode"]; ok {
			gameMode = mode.(string)
		}

		// Verify all entries have the same game mode
		for i, entry := range entries {
			entryMode := "classic"
			if mode, ok := entry.GetProperties()["mode"]; ok {
				entryMode = mode.(string)
			}
			if entryMode != gameMode {
				logger.Error("Matchmaker entries have different game modes: entry %d has %s, expected %s", i, entryMode, gameMode)
				return "", runtime.NewError("mismatched game modes", 3)
			}
		}

		logger.Info("Creating match for game mode: %s with %d players", gameMode, len(entries))
		matchLabel := "lobby_" + gameMode
		matchId, err := nk.MatchCreate(ctx, matchLabel, map[string]interface{}{"mode": gameMode, "invited": entries})
		if err != nil {
			return "", err
		}
		return matchId, nil
	}); err != nil {
		logger.Error("unable to register matchmaker matched hook: %v", err)
		return err
	}

	// Register match handlers for each game mode
	if err := initializer.RegisterMatch("lobby_classic", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		logger.Info("=== CREATING NEW CLASSIC MATCH INSTANCE ===")
		match := NewMatchWithMode("classic")
		logger.Info("=== CLASSIC MATCH INSTANCE CREATED SUCCESSFULLY ===")
		return match, nil
	}); err != nil {
		logger.Error("unable to register classic match: %v", err)
		return err
	}

	if err := initializer.RegisterMatch("lobby_timed", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		logger.Info("=== CREATING NEW TIMED MATCH INSTANCE ===")
		match := NewMatchWithMode("timed")
		logger.Info("=== TIMED MATCH INSTANCE CREATED SUCCESSFULLY ===")
		return match, nil
	}); err != nil {
		logger.Error("unable to register timed match: %v", err)
		return err
	}

	// Keep the generic lobby handler for backward compatibility
	if err := initializer.RegisterMatch("lobby", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		logger.Info("=== CREATING NEW MATCH INSTANCE ===")
		match := NewMatch()
		logger.Info("=== MATCH INSTANCE CREATED SUCCESSFULLY ===")
		return match, nil
	}); err != nil {
		logger.Error("unable to register: %v", err)
		return err
	}

	id := "TicTacToeLeaderboard"
	authoritative := true
	sort := "desc"
	operator := "best"
	reset := "0 0 * * 1"

	if err := nk.LeaderboardCreate(ctx, id, authoritative, sort, operator, reset, nil, false); err != nil {
		logger.Error("unable to create leaderboard: %v", err)
	}

	// Register RPC for top N leaderboard
	if err := initializer.RegisterRpc("GetTopPlayers", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
		// Parse N from payload (default 10)
		n := 10
		if payload != "" {
			var req struct {
				N int `json:"n"`
			}
			if err := json.Unmarshal([]byte(payload), &req); err == nil && req.N > 0 {
				n = req.N
			}
		}
		// Fetch top N records
		records, _, _, _, err := nk.LeaderboardRecordsList(ctx, "TicTacToeLeaderboard", nil, n, "", 0)
		if err != nil {
			logger.Error("LeaderboardRecordsList error: %v", err)
			return "{}", errInternal
		}
		// Prepare response
		type Player struct {
			Username string `json:"username"`
			OwnerId  string `json:"owner_id"`
			Score    int64  `json:"score"`
		}
		var players []Player
		for _, r := range records {
			players = append(players, Player{
				Username: r.GetUsername().GetValue(),
				OwnerId:  r.GetOwnerId(),
				Score:    r.GetScore(),
			})
		}
		respBytes, _ := json.Marshal(players)
		return string(respBytes), nil
	}); err != nil {
		logger.Error("unable to register GetTopPlayers RPC: %v", err)
		return err
	}

	return nil
}
