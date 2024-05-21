package main

import (
	"os"

	"github.com/saizo80/mcrcon-server/gen"
	"github.com/saizo80/mcrcon-server/log"
	serverfilefunctions "github.com/saizo80/mcrcon-server/serverFileFunctions"
)

func main() {
	args := os.Args[1:]

	if gen.Contains(args, "-d", "--debug") {
		log.DebugMode = true
		log.Debug("debug logging enabled")
	}

	log.Info("starting mcrcon-server")
	err := gen.GenInit()
	if err != nil {
		os.Exit(1)
	}
	defer gen.DB.DB.Close()
	defer gen.Conn.Conn.Close()
	err = serverfilefunctions.InitPlayers()
	if err != nil {
		os.Exit(1)
	}
	err = serverfilefunctions.InitServerProperties()
	if err != nil {
		os.Exit(1)
	}
}
