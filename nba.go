package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Row interface {
	Scan(dest ...interface{}) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
}

type Cache interface {
	Get(key string) (string, error)
	Set(key string, value interface{}) error
	Del(key string) error
	Close()
}

type Database interface {
	Exec(query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) Row
	Query(query string, args ...interface{}) (Rows, error)
	Close()
}

type PlayerRecord struct {
	ID        int     `json:"Id"`
	Points    int     `json:"points"`
	Rebounds  int     `json:"rebounds"`
	Assists   int     `json:"assists"`
	Steals    int     `json:"steals"`
	Blocks    int     `json:"blocks"`
	Turnovers int     `json:"turnovers"`
	Fouls     int     `json:"fouls"`
	Minutes   float64 `json:"minutes"`
}

type AggregatedRecord struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Points    float64 `json:"points"`
	Rebounds  float64 `json:"rebounds"`
	Assists   float64 `json:"assists"`
	Steals    float64 `json:"steals"`
	Blocks    float64 `json:"blocks"`
	Turnovers float64 `json:"turnovers"`
	Fouls     float64 `json:"fouls"`
	Minutes   float64 `json:"minutes"`
}

type NBAStatistics struct {
	cache Cache
	db    Database
}

func NewNBAStatistics(cache Cache, db Database) *NBAStatistics {
	return &NBAStatistics{
		cache: cache,
		db:    db,
	}
}

func NewPlayerRecordFromJSON(data io.ReadCloser) (*PlayerRecord, error) {
	var record PlayerRecord
	err := json.NewDecoder(data).Decode(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (record *PlayerRecord) Validate() error {
	if record.Fouls > 6 {
		return fmt.Errorf("fouls cannot be greater than 6")
	}
	if record.Minutes < 0 || record.Minutes > 48.0 {
		return fmt.Errorf("minutes must be between 0 and 48")
	}
	if record.Points < 0 || record.Rebounds < 0 || record.Assists < 0 || record.Steals < 0 || record.Blocks < 0 || record.Turnovers < 0 {
		return fmt.Errorf("statistics values cannot be negative")
	}
	return nil
}

func (nba *NBAStatistics) addPlayerRecord(w http.ResponseWriter, r *http.Request) {
	// Create and validate PlayerRecord
	record, err := NewPlayerRecordFromJSON(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := record.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert record into db
	err = nba.db.Exec("INSERT INTO player_records (player_id, points, rebounds, assists, steals, blocks, turnovers, fouls, minutes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		record.ID, record.Points, record.Rebounds, record.Assists, record.Steals, record.Blocks, record.Turnovers, record.Fouls, record.Minutes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Invalidate cache
	nba.cache.Del(fmt.Sprintf("player_aggregate_%d", record.ID))
	teamID, err := nba.getTeamIDByPlayerID(record.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nba.cache.Del("team_aggregate_%d" + strconv.Itoa(teamID))

	w.WriteHeader(http.StatusCreated)
}

func (nba *NBAStatistics) getPlayerAggregate(w http.ResponseWriter, r *http.Request) {
	playerIDStr := r.URL.Query().Get("playerId")
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		http.Error(w, "Invalid playerId", http.StatusBadRequest)
		return
	}

	// Check cache first
	cacheKey := fmt.Sprintf("player_aggregate_%d", playerID)
	cachedResult, err := nba.cache.Get(cacheKey)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResult))
		return
	}

	// Query db for aggregate data
	var aggregate AggregatedRecord
	err = nba.db.QueryRow(`
		SELECT pr.player_id, p.name, AVG(pr.points), AVG(pr.rebounds), AVG(pr.assists), AVG(pr.steals), AVG(pr.blocks), AVG(pr.turnovers), AVG(pr.fouls), AVG(pr.minutes)
		FROM player_records pr
		JOIN player p ON pr.player_id = p.id
		WHERE pr.player_id=$1
		GROUP BY pr.player_id, p.name`, playerID).Scan(
		&aggregate.ID, &aggregate.Name, &aggregate.Points, &aggregate.Rebounds, &aggregate.Assists, &aggregate.Steals, &aggregate.Blocks, &aggregate.Turnovers, &aggregate.Fouls, &aggregate.Minutes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the result
	resultJSON, _ := json.Marshal(aggregate)
	err = nba.cache.Set(cacheKey, resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func (nba *NBAStatistics) getTeamAggregate(w http.ResponseWriter, r *http.Request) {
	teamIDStr := r.URL.Query().Get("teamId")
	teamID, err := strconv.Atoi(teamIDStr)
	if err != nil {
		http.Error(w, "Invalid teamId", http.StatusBadRequest)
		return
	}

	// Check cache first
	cacheKey := fmt.Sprintf("team_aggregate_%d", teamID)
	cachedResult, err := nba.cache.Get(cacheKey)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cachedResult))
		return
	}

	// Query db for aggregate data
	var aggregate AggregatedRecord
	err = nba.db.QueryRow(`
		SELECT p.team_id, t.name, AVG(pr.points), AVG(pr.rebounds), AVG(pr.assists), AVG(pr.steals), AVG(pr.blocks), AVG(pr.turnovers), AVG(pr.fouls), AVG(pr.minutes)
		FROM player_records pr
		JOIN player p ON pr.player_id = p.player_id
		JOIN team t ON p.team_id = t.team_id
		WHERE p.team_id = $1
		GROUP BY t.name`, teamID).Scan(
		&aggregate.ID, &aggregate.Name, &aggregate.Points, &aggregate.Rebounds, &aggregate.Assists, &aggregate.Steals, &aggregate.Blocks, &aggregate.Turnovers, &aggregate.Fouls, &aggregate.Minutes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache the result
	resultJSON, _ := json.Marshal(aggregate)
	err = nba.cache.Set(cacheKey, resultJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func (nba *NBAStatistics) getAllPlayersAggregate(w http.ResponseWriter, r *http.Request) {
	// Query db for aggregate data
	rows, err := nba.db.Query(`
		SELECT p.player_id, AVG(pr.points), AVG(pr.rebounds), AVG(pr.assists), AVG(pr.steals), AVG(pr.blocks), AVG(pr.turnovers), AVG(pr.fouls), AVG(pr.minutes), p.name
		FROM player_records pr
		JOIN player p ON pr.player_id = p.player_id
		GROUP BY p.player_id, p.name`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []AggregatedRecord

	for rows.Next() {
		var record AggregatedRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Points, &record.Rebounds, &record.Assists, &record.Steals, &record.Blocks, &record.Turnovers, &record.Fouls, &record.Minutes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	resultJSON, _ := json.Marshal(records)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func (nba *NBAStatistics) getAllTeamsAggregate(w http.ResponseWriter, r *http.Request) {
	// Query db for aggregate data
	rows, err := nba.db.Query(`
		SELECT t.team_id, AVG(pr.points), AVG(pr.rebounds), AVG(pr.assists), AVG(pr.steals), AVG(pr.blocks), AVG(pr.turnovers), AVG(pr.fouls), AVG(pr.minutes), t.name
		FROM player_records pr
		JOIN player p ON pr.player_id = p.player_id
		JOIN team t ON p.team_id = t.team_id
		GROUP BY t.team_id, t.name`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []AggregatedRecord

	for rows.Next() {
		var record AggregatedRecord
		if err := rows.Scan(&record.ID, &record.Name, &record.Points, &record.Rebounds, &record.Assists, &record.Steals, &record.Blocks, &record.Turnovers, &record.Fouls, &record.Minutes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		records = append(records, record)
	}

	resultJSON, _ := json.Marshal(records)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func (nba *NBAStatistics) getTeamIDByPlayerID(playerID int) (int, error) {
	cacheKey := fmt.Sprintf("player_team_%d", playerID)
	cachedTeamID, err := nba.cache.Get(cacheKey)
	if err == nil {
		teamID, err := strconv.Atoi(cachedTeamID)
		if err == nil {
			return teamID, nil
		}
	}

	var teamID int
	err = nba.db.QueryRow("SELECT team_id FROM player WHERE player_id=$1", playerID).Scan(&teamID)
	if err != nil {
		return 0, err
	}

	err = nba.cache.Set(cacheKey, teamID)
	if err != nil {
		return 0, err
	}

	return teamID, nil
}
