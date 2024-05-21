package gen

import (
	"github.com/saizo80/mcrcon-server/mcSql"
	"github.com/saizo80/mcrcon-server/mcrcon"
)

var (
	DB       mcSql.SQLInterface
	Conn     mcrcon.Rcon
	Server   string
	Port     string
	Password string
)

const (
	MCDataDirectory = "./mcTestData/mcdata"
)

func GenInit() error {
	err := Conn.Init()
	if err != nil {
		return err
	}
	err = DB.Init()
	if err != nil {
		return err
	}
	return nil
}

// check if the slice contains any of the given strings
func Contains(s []string, c ...string) bool {
	for _, v := range s {
		for _, check := range c {
			if v == check {
				return true
			}
		}
	}
	return false
}
