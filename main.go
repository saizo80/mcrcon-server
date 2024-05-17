package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gorcon/rcon"
)

func main() {
	server := os.Getenv("RCON_SERVER")
	port := os.Getenv("RCON_PORT")
	password := os.Getenv("RCON_PASSWORD")
	if server == "" || port == "" || password == "" {
		fmt.Println("Please set RCON_SERVER, RCON_PORT, and RCON_PASSWORD environment variables")
		os.Exit(1)
	}

	conn, err := rcon.Dial(server+":"+port, password)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	response, err := conn.Execute("help")
	if err != nil {
		panic(err)
	}

	// split the response into lines
	lines := strings.Split(response, "/")[1:]
	for _, line := range lines {
		fmt.Println("/" + line)
	}
}
