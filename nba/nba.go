package nba

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ShimonMoldawskiy/NBAStatistics/common"
)

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value interface{}) error
	Del(key string) error
	Close()
}

type Database interface {
	Exec(query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) common.Row
	Query(query string, args ...interface{}) (common.Rows, error)
	Close()
}

type NBAStatistics struct {
	cache   Cache
	db      Database
	teams   map[int]Team
	players map[int]Player
}

func NewNBAStatistics(cache Cache, db Database) (*NBAStatistics, error) {
	teams, err := GetTeams(db)
	if err != nil {
		return nil, err
	}
	players, err := GetPlayers(db, teams)
	if err != nil {
		return nil, err
	}
	return &NBAStatistics{
		cache:   cache,
		db:      db,
		teams:   teams,
		players: players,
	}, nil
}

func (nba *NBAStatistics) AddRecord(w http.ResponseWriter, r *http.Request) {
	// Create and validate Record
	record, err := NewRecord(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var (
		player Player
		exists bool
	)
	if player, exists = nba.players[record.ID]; !exists {
		http.Error(w, fmt.Sprintf("player with ID %d does not exist", record.ID), http.StatusBadRequest)
		return
	}

	if err := record.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert record into db
	err = record.saveToDB(nba.db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Invalidate cache
	if err = nba.cache.Del(player.CacheKey()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err = nba.cache.Del(player.Team.CacheKey()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
}

func (nba *NBAStatistics) getAggregateData(a AggregatedObject) ([]byte, error) {
	// Check cache first
	cachedResult, err := nba.cache.Get(a.CacheKey())
	if err == nil {
		return []byte(cachedResult), nil
	}

	// Query db for aggregate data
	var aggregate *AggregatedRecord = a.NewAggregatedRecord()
	if err = nba.db.QueryRow(a.DBQuery()).Scan(
		aggregate.Points, aggregate.Rebounds, aggregate.Assists, aggregate.Steals, aggregate.Blocks,
		aggregate.Turnovers, aggregate.Fouls, aggregate.Minutes); err != nil {
		return nil, err
	}

	result, err := json.Marshal(*aggregate)
	if err != nil {
		return nil, err
	}

	// Put the result to cache
	err = nba.cache.Set(a.CacheKey(), result)

	return result, err
}

func (nba *NBAStatistics) GetPlayerAggregate(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.URL.Query().Get("playerId")
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		http.Error(w, "Invalid playerId", http.StatusBadRequest)
		return
	}

	var (
		player Player
		exists bool
	)
	if player, exists = nba.players[playerID]; !exists {
		http.Error(w, fmt.Sprintf("player with ID %d does not exist", playerID), http.StatusBadRequest)
		return
	}

	result, err := nba.getAggregateData(player)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)

}

func (nba *NBAStatistics) GetTeamAggregate(w http.ResponseWriter, r *http.Request) {
	teamIDStr := r.URL.Query().Get("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		http.Error(w, "Invalid teamId", http.StatusBadRequest)
		return
	}

	var (
		team   Team
		exists bool
	)
	if team, exists = nba.teams[teamID]; !exists {
		http.Error(w, fmt.Sprintf("team with ID %d does not exist", teamID), http.StatusBadRequest)
		return
	}

	result, err := nba.getAggregateData(team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)

}

func (nba *NBAStatistics) GetAllPlayersAggregate(w http.ResponseWriter, r *http.Request) {
	var records []AggregatedRecord

	for _, player := range nba.players {
		result, err := nba.getAggregateData(player)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var record AggregatedRecord
		if err := json.Unmarshal(result, &record); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	resultJSON, _ := json.Marshal(records)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func (nba *NBAStatistics) GetAllTeamsAggregate(w http.ResponseWriter, r *http.Request) {
	var records []AggregatedRecord

	for _, team := range nba.teams {
		result, err := nba.getAggregateData(team)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var record AggregatedRecord
		if err := json.Unmarshal(result, &record); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	resultJSON, _ := json.Marshal(records)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}
