package mcSql

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/saizo80/mcrcon-server/log"
)

const (
	databaseFile = "./database.db"
)

type Player struct {
	UUID        string
	Name        string
	LastSeen    string
	Active      bool
	Op          bool
	Banned      bool
	Whitelisted bool
}

func (p Player) String() string {
	return fmt.Sprintf(
		"UUID: %s, Name: %s, LastSeen: %s, Active: %t, Op: %t, Banned: %t, Whitelisted: %t",
		p.UUID, p.Name, p.LastSeen, p.Active, p.Op, p.Banned, p.Whitelisted,
	)
}

type SQLInterface struct {
	// db connection
	DB *sql.DB
}

func (s *SQLInterface) Init() error {
	log.Debug("opening database connection")
	// connect to the database
	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return log.Error(err)
	}
	s.DB = db
	return s.setup()
}

func (s *SQLInterface) setup() error {
	// create the players table
	log.Debug("creating players table")
	_, err := s.DB.Exec(`create table if not exists players (
		uuid text primary key,
		name text,
		last_seen text,
		active boolean,
		op boolean,
		banned boolean,
		whitelisted boolean
	)`)
	if err != nil {
		return log.Error(err)
	}
	_, err = s.DB.Exec(`create table if not exists server_properties (
		key text,
		value text,
		revision integer default 0,
		date text,
		active boolean default true,
		primary key (key, revision)
	)`)
	if err != nil {
		return log.Error(err)
	}
	return nil
}

func (s *SQLInterface) UpdatePropery(key string, value string) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return log.Error(err)
	}
	result, err := tx.Query("select coalesce(max(revision), 0) from server_properties where key = ?", key)
	if err != nil {
		tx.Rollback()
		return log.Error(err)
	}
	var revision int
	for result.Next() {
		result.Scan(&revision)
	}
	revision++
	_, err = tx.Exec("update server_properties set active = false where key = ?", key)
	if err != nil {
		tx.Rollback()
		return log.Error(err)
	}
	_, err = tx.Exec("insert into server_properties (key, value, revision, date) values (?, ?, ?, datetime('now'))", key, value, revision)
	if err != nil {
		tx.Rollback()
		return log.Error(err)
	}
	tx.Commit()
	return nil
}

func (s *SQLInterface) GetPlayers() ([]Player, error) {
	rows := s.Query("select * from players")
	players := []Player{}
	for rows.Next() {
		var uuid string
		var name string
		var last_seen string
		var active bool
		var op bool
		var banned bool
		var whitelisted bool
		err := rows.Scan(&uuid, &name, &last_seen, &active, &op, &banned, &whitelisted)
		if err != nil {
			return nil, log.Error(err)
		}
		players = append(players, Player{uuid, name, last_seen, active, op, banned, whitelisted})
	}
	return players, nil
}

func (s *SQLInterface) InsertPlayer(player Player) {
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}
	exec := `insert or ignore into players (
		uuid, name, last_seen, active, op, banned, whitelisted
	) values (?, ?, ?, ?, ?, ?, ?)`
	tx.Exec(exec, player.UUID, player.Name, player.LastSeen, player.Active, player.Op, player.Banned, player.Whitelisted)
	tx.Commit()
}

func (s *SQLInterface) UpdatePlayerActive(uuid string, active bool) error {
	tx, err := s.DB.Begin()
	if err != nil {
		tx.Rollback()
		return log.Error(err)
	}
	_, err = tx.Exec("update players set active = ? where (uuid = ? or name = ?)", active, uuid, uuid)
	if err != nil {
		tx.Rollback()
		return log.Error(err)
	}
	tx.Commit()
	return nil
}

func (s *SQLInterface) UpdatePlayerOp(uuid string, op bool) {
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}
	tx.Exec("update players set op = ? where uuid = ?", op, uuid)
	tx.Commit()
}

func (s *SQLInterface) UpdatePlayerBanned(uuid string, banned bool) {
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}
	tx.Exec("update players set banned = ? where uuid = ?", banned, uuid)
	tx.Commit()
}

func (s *SQLInterface) UpdatePlayerWhitelisted(uuid string, whitelisted bool) {
	tx, err := s.DB.Begin()
	if err != nil {
		panic(err)
	}
	tx.Exec("update players set whitelisted = ? where uuid = ?", whitelisted, uuid)
	tx.Commit()
}

func (s *SQLInterface) Query(query string) *sql.Rows {
	rows, err := s.DB.Query(query)
	if err != nil {
		panic(err)
	}
	return rows
}

func (s *SQLInterface) Exec(stmt string, args ...interface{}) (sql.Result, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(stmt, args...)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return result, nil
}
