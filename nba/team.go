package nba

import (
	"fmt"
)

type Team struct {
	ID   int
	Name string
}

func GetTeams(db Database) (map[int]Team, error) {
	rows, err := db.Query("SELECT id, name FROM teams")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make(map[int]Team)
	for rows.Next() {
		var team Team
		if err := rows.Scan(&team.ID, &team.Name); err != nil {
			return nil, err
		}
		teams[team.ID] = team
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return teams, nil
}

func (t Team) NewAggregatedRecord() *AggregatedRecord {
	return &AggregatedRecord{ID: t.ID, Name: t.Name}
}

func (t Team) CacheKey() string {
	return fmt.Sprintf("team_%d", t.ID)
}

func (t Team) DBQuery() string {
	return fmt.Sprintf(`SELECT AVG(r.points) AS points, AVG(r.rebounds) AS rebounds, AVG(r.assists) AS assists, AVG(r.steals) AS steals, AVG(r.blocks) AS blocks, AVG(r.turnovers) AS turnovers, AVG(r.fouls) AS fouls, AVG(r.minutes) AS minutes FROM records r JOIN players p ON r.player_id = p.id WHERE p.team_id=%d;`, t.ID)
}
