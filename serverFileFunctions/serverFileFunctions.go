package serverfilefunctions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saizo80/mcrcon-server/gen"
	"github.com/saizo80/mcrcon-server/log"
	"github.com/saizo80/mcrcon-server/mcSql"
)

type jsonPlayers struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

func InitPlayers() error {
	log.Debug("getting most current player data from server files")
	// read current players from whitelist.json, banned-players.json, ops.json, and usercache.json
	// to get the UUIDs and names and insert them into the database
	baseDir := gen.MCDataDirectory
	whitelist := baseDir + "/whitelist.json"
	banned := baseDir + "/banned-players.json"
	ops := baseDir + "/ops.json"
	usercache := baseDir + "/usercache.json"

	usercachePlayers := readFile(usercache)
	for _, player := range usercachePlayers {
		gen.DB.InsertPlayer(mcSql.Player{
			UUID:        player.UUID,
			Name:        player.Name,
			Active:      false,
			Op:          false,
			Banned:      false,
			Whitelisted: false,
		})
	}

	_, err := gen.DB.Exec("update players set whitelisted = false")
	if err != nil {
		return log.Error(err)
	}
	whitelistPlayers := readFile(whitelist)
	whitelistUUIDs := "("
	for _, player := range whitelistPlayers {
		whitelistUUIDs += fmt.Sprintf("'%s',", player.UUID)
	}
	whitelistUUIDs = whitelistUUIDs[:len(whitelistUUIDs)-1] + ")"
	_, err = gen.DB.Exec(fmt.Sprintf("update players set whitelisted = true where uuid in %s", whitelistUUIDs))
	if err != nil {
		return log.Error(err)
	}

	_, err = gen.DB.Exec("update players set banned = false")
	if err != nil {
		return log.Error(err)
	}
	bannedPlayers := readFile(banned)
	bannedUUIDs := "("
	for _, player := range bannedPlayers {
		bannedUUIDs += fmt.Sprintf("'%s',", player.UUID)
	}
	bannedUUIDs = bannedUUIDs[:len(bannedUUIDs)-1] + ")"
	_, err = gen.DB.Exec(fmt.Sprintf("update players set banned = true where uuid in %s", bannedUUIDs))
	if err != nil {
		return log.Error(err)
	}

	_, err = gen.DB.Exec("update players set op = false")
	if err != nil {
		return log.Error(err)
	}
	opsPlayers := readFile(ops)
	opsUUIDs := "("
	for _, player := range opsPlayers {
		opsUUIDs += fmt.Sprintf("'%s',", player.UUID)
	}
	opsUUIDs = opsUUIDs[:len(opsUUIDs)-1] + ")"
	_, err = gen.DB.Exec(fmt.Sprintf("update players set op = true where uuid in %s", opsUUIDs))
	if err != nil {
		return log.Error(err)
	}

	_, err = gen.DB.Exec("update players set active = false")
	if err != nil {
		return log.Error(err)
	}

	response, err := gen.Conn.Execute("list")
	if err != nil {
		return log.Error(err)
	}
	if !strings.Contains(response, "There are 0 of a max of") {
		step := strings.Split(response, ":")
		players := strings.Split(step[1], ",")
		log.Debug("current number of players online: %v", players)
		for _, player := range players {
			err := gen.DB.UpdatePlayerActive(strings.TrimSpace(player), true)
			if err != nil {
				return err
			}
		}
	} else {
		log.Debug("no players online currently")
	}

	return nil
}

func readFile(file string) []jsonPlayers {
	// read the file and return a slice of jsonPlayers
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := io.Reader(f)
	decoder := json.NewDecoder(reader)
	players := []jsonPlayers{}
	err = decoder.Decode(&players)
	if err != nil {
		panic(err)
	}
	return players
}
