package mcrcon

import (
	"github.com/gorcon/rcon"
	"github.com/saizo80/mcrcon-server/log"
)

type Rcon struct {
	Conn *rcon.Conn
}

func (r *Rcon) Init(server string, port string, password string) error {
	log.Debug("connecting to rcon on %s:%s", server, port)
	conn, err := rcon.Dial(server+":"+port, password)
	r.Conn = conn
	return log.Error(err)
}

func (r *Rcon) Execute(command string) (string, error) {
	return r.Conn.Execute(command)
}
