package main

// This is the main package for the NBA Statistics project.
// The project is located at "github.com/ShimonMoldawskiy/NBAStatistics".

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/ShimonMoldawskiy/NBAStatistics/cache"
	"github.com/ShimonMoldawskiy/NBAStatistics/db"
	"github.com/ShimonMoldawskiy/NBAStatistics/nba"
)

func main() {
	ctx, cfn := context.WithCancelCause(context.Background())
	defer cfn(nil)

	defer func() {
		if err := recover(); err != nil {
			err := fmt.Errorf("panic in main, err: %v", err)
			cfn(err)
			log.Fatal(ctx, err)
		}
	}()

	// Initialize db connection, perform migrations if necessary
	dbHost := os.Getenv("POSTGRES_HOST")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	if dbHost == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Postgres environment variables are not set")
	}

	var err error
	connString := fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbName)
	dbMigration, err := goose.OpenDBWithDriver("postgres", connString)
	if err != nil {
		log.Fatalf("Unable to migrate the database: %v\n", err)
	}
	dbMigration.Close()
	db, err := db.NewPostgresDatabase(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	// Initialize cache connection
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatal("REDIS_HOST environment variable is not set")
	}
	cache, err := cache.NewRedisCache(ctx, redisHost+":6379", "", 0)
	if err != nil {
		log.Fatalf("Unable to connect to cache: %v\n", err)
	}
	defer cache.Close()

	// Initialize NBAStatistics
	nba, err := nba.NewNBAStatistics(cache, db)
	if err != nil {
		log.Fatalf("Unable to initialize statistics: %v\n", err)
	}

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/record", nba.AddRecord).Methods("POST")
	r.HandleFunc("/aggregate/player", nba.GetPlayerAggregate).Methods("GET")
	r.HandleFunc("/aggregate/team", nba.GetTeamAggregate).Methods("GET")
	r.HandleFunc("/aggregate/players", nba.GetAllPlayersAggregate).Methods("GET")
	r.HandleFunc("/aggregate/teams", nba.GetAllTeamsAggregate).Methods("GET")

	// Start server
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
