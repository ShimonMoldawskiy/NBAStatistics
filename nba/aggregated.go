package nba

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

type AggregatedObject interface {
	NewAggregatedRecord() *AggregatedRecord
	CacheKey() string
	DBQuery() string
}
