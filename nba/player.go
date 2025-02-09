package nba

import (
	"fmt"
)

type Player struct {
	ID   int
	Name string
	Team Team
}

func GetPlayers(db Database, teams map[int]Team) (map[int]Player, error) {
	rows, err := db.Query("SELECT id, name, team_id FROM players")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make(map[int]Player)
	for rows.Next() {
		var player Player
		var teamID int
		if err := rows.Scan(&player.ID, &player.Name, &teamID); err != nil {
			return nil, err
		}
		team, ok := teams[teamID]
		if !ok {
			return nil, fmt.Errorf("team with ID %d not found", teamID)
		}
		player.Team = team

		players[player.ID] = player
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}

func (p Player) NewAggregatedRecord() *AggregatedRecord {
	return &AggregatedRecord{ID: p.ID, Name: p.Name}
}

func (p Player) CacheKey() string {
	return fmt.Sprintf("player_%d", p.ID)
}

func (p Player) DBQuery() string {
	return fmt.Sprintf(`SELECT AVG(r.points), AVG(r.rebounds), AVG(r.assists), AVG(r.steals), AVG(r.blocks), AVG(r.turnovers), AVG(r.fouls), AVG(r.minutes)
		FROM records r
		WHERE r.player_id=%d`, p.ID)
}
