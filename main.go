package main

// This is the main package for the NBA Statistics project.
// The project is located at "github.com/ShimonMoldawskiy/NBAStatistics".

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/ShimonMoldawskiy/NBAStatistics/cache"
	"github.com/ShimonMoldawskiy/NBAStatistics/db"
)

var (
	ctx = context.Background()
)

func main() {
	// Initialize db connection
	var err error
	db, err := db.NewPostgresDatabase(ctx, "postgresql://user:password@primary-db-host/dbname")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	// Initialize cache connection
	cache, err := cache.NewRedisCache(ctx, "redis-host:6379", "", 0)
	if err != nil {
		log.Fatalf("Unable to connect to cache: %v\n", err)
	}
	defer cache.Close()

	// Initialize NBAStatistics
	nba := NewNBAStatistics(cache, db)

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/record", nba.addPlayerRecord).Methods("POST")
	r.HandleFunc("/aggregate/player", nba.getPlayerAggregate).Methods("GET")
	r.HandleFunc("/aggregate/team", nba.getTeamAggregate).Methods("GET")
	r.HandleFunc("/aggregate/players", nba.getAllPlayersAggregate).Methods("GET")
	r.HandleFunc("/aggregate/teams", nba.getAllTeamsAggregate).Methods("GET")

	// Start server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
