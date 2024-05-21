package mcrcon

import (
	"os"
	"strings"
	"time"

	"github.com/gorcon/rcon"
	"github.com/saizo80/mcrcon-server/log"
)

const (
	maxRetries = 3
)

type Rcon struct {
	Conn *rcon.Conn
}

func (r *Rcon) Init() error {
	server, port, password := getEnvVars()
	log.Debug("connecting to rcon on %s:%s", server, port)
	counter := 0
	for counter < maxRetries {
		conn, err := rcon.Dial(server+":"+port, password)
		if err == nil {
			r.Conn = conn
			return nil
		}
		counter++
		log.Warn("failed to connect to rcon, retrying in 5 seconds (%d/%d)", counter, maxRetries)
		time.Sleep(5 * time.Second)
	}
	return log.Errorf("failed to connect to rcon after %d retries", maxRetries)
}

func (r *Rcon) Execute(command string) (string, error) {
	err := r.ping()
	if err != nil {
		return "", err
	}
	return r.Conn.Execute(command)
}

func (r *Rcon) ping() error {
	result, err := r.Conn.Execute("list")
	if err != nil || !strings.Contains(result, "There are") {
		err = r.reconnect()
	}
	return err
}

func (r *Rcon) reconnect() error {
	r.Conn.Close()
	return r.Init()
}

func getEnvVars() (string, string, string) {
	server := os.Getenv("RCON_SERVER")
	port := os.Getenv("RCON_PORT")
	password := os.Getenv("RCON_PASSWORD")
	if server == "" || port == "" || password == "" {
		log.Errorf("RCON_SERVER, RCON_PORT, and RCON_PASSWORD must be set")
		os.Exit(1)
	}
	return server, port, password
}
