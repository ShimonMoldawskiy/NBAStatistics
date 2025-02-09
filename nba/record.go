package nba

import (
	"encoding/json"
	"fmt"
	"io"
)

type Record struct {
	ID        int     `json:"id"`
	Points    int     `json:"points"`
	Rebounds  int     `json:"rebounds"`
	Assists   int     `json:"assists"`
	Steals    int     `json:"steals"`
	Blocks    int     `json:"blocks"`
	Turnovers int     `json:"turnovers"`
	Fouls     int     `json:"fouls"`
	Minutes   float64 `json:"minutes"`
}

func NewRecord(data io.ReadCloser) (*Record, error) {
	var record Record
	err := json.NewDecoder(data).Decode(&record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (record *Record) Validate() error {
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

func (record *Record) saveToDB(db Database) error {
	return db.Exec("INSERT INTO records (player_id, points, rebounds, assists, steals, blocks, turnovers, fouls, minutes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		record.ID, record.Points, record.Rebounds, record.Assists, record.Steals, record.Blocks, record.Turnovers, record.Fouls, record.Minutes)
}
