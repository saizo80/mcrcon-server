package mcSql

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Player struct {
	UUID     string
	Name     string
	LastSeen string
	Active   bool
}

type SQLInterface struct {
	// db connection
	DB *sql.DB
}

func (s *SQLInterface) Init() SQLInterface {
	// connect to the database
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		panic(err)
	}
	s.DB = db
	s.setup()
	return *s
}

func (s *SQLInterface) setup() {
	// create the players table
	s.DB.Exec(`create table if not exists players (
		uuid text primary key,
		name text,
		last_seen text,
		active boolean
	)`)

}

func (s *SQLInterface) GetPlayers() []Player {
	rows := s.Query("select * from players")
	players := []Player{}
	for rows.Next() {
		var uuid string
		var name string
		var last_seen string
		var active bool
		rows.Scan(&uuid, &name, &last_seen, &active)
		players = append(players, Player{uuid, name, last_seen, active})
	}
	return players
}

func (s *SQLInterface) InsertPlayer(player Player) {
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}
	tx.Exec("insert into players (uuid, name, last_seen, active) values (?, ?, ?, ?)", player.UUID, player.Name, player.LastSeen, player.Active)
	tx.Commit()
}

func (s *SQLInterface) UpdatePlayerActive(uuid string, active bool) {
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}
	tx.Exec("update players set active = ? where uuid = ?", active, uuid)
	tx.Commit()
}

func (s *SQLInterface) Query(query string) *sql.Rows {
	rows, err := s.DB.Query(query)
	if err != nil {
		panic(err)
	}
	return rows
}
