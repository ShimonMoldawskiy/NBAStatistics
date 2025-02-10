-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO teams (name) VALUES ('Lakers');
INSERT INTO teams (name) VALUES ('Warriors');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE players (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    team_id INTEGER REFERENCES teams(id)
);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO players (name, team_id) VALUES ('Lebron James', 1);
INSERT INTO players (name, team_id) VALUES ('Anthony Davis', 1);
INSERT INTO players (name, team_id) VALUES ('Stephen Curry', 2);
INSERT INTO players (name, team_id) VALUES ('Klay Thompson', 2);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    player_id INTEGER REFERENCES players(id),
	points    INTEGER,
	rebounds  INTEGER,
	assists   INTEGER,
	steals    INTEGER,
	blocks    INTEGER,
	turnovers INTEGER,
	fouls     INTEGER,
    minutes   FLOAT
    );
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS records;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS players;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd