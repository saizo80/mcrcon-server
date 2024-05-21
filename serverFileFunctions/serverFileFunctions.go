package serverfilefunctions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saizo80/mcrcon-server/gen"
	"github.com/saizo80/mcrcon-server/log"
	"github.com/saizo80/mcrcon-server/mcSql"
)

func InitPlayers() error {
	log.Debug("getting most current player data from server files")
	// read current players from whitelist.json, banned-players.json, ops.json, and usercache.json
	// to get the UUIDs and names and insert them into the database
	baseDir := gen.MCDataDirectory
	whitelist := baseDir + "/whitelist.json"
	banned := baseDir + "/banned-players.json"
	ops := baseDir + "/ops.json"
	usercache := baseDir + "/usercache.json"

	usercachePlayers, err := readPlayerFile(usercache)
	if err != nil {
		return err
	}
	for _, player := range usercachePlayers {
		gen.DB.InsertPlayer(mcSql.Player{
			UUID:        player["uuid"].(string),
			Name:        player["name"].(string),
			Active:      false,
			Op:          false,
			Banned:      false,
			Whitelisted: false,
		})
	}

	_, err = gen.DB.Exec("update players set whitelisted = false")
	if err != nil {
		return log.Error(err)
	}
	whitelistPlayers, err := readPlayerFile(whitelist)
	if err != nil {
		return log.Error(err)
	}
	whitelistUUIDs := "("
	for _, player := range whitelistPlayers {
		whitelistUUIDs += fmt.Sprintf("'%s',", player["uuid"])
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
	bannedPlayers, err := readPlayerFile(banned)
	if err != nil {
		return log.Error(err)
	}
	for _, player := range bannedPlayers {
		err := gen.DB.UpdatePlayerBanned(player["uuid"].(string), true, player["reason"].(string))
		if err != nil {
			return log.Error(err)
		}

	}

	_, err = gen.DB.Exec("update players set op = false")
	if err != nil {
		return log.Error(err)
	}
	opsPlayers, err := readPlayerFile(ops)
	if err != nil {
		return log.Error(err)
	}
	opsUUIDs := "("
	for _, player := range opsPlayers {
		opsUUIDs += fmt.Sprintf("'%s',", player["uuid"])
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
				return log.Error(err)
			}
		}
	} else {
		log.Debug("no players online currently")
	}

	return nil
}

func InitServerProperties() error {
	log.Debug("getting most current server properties from server.properties file")
	// read the server.properties file and insert the properties into the database
	baseDir := gen.MCDataDirectory
	serverProperties := baseDir + "/server.properties"
	f, err := os.Open(serverProperties)
	if err != nil {
		return log.Error(err)
	}
	defer f.Close()

	reader := io.Reader(f)

	// read the server.properties file
	properties := map[string]string{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.Split(line, "=")
			properties[parts[0]] = parts[1]
		}
	}

	// insert the properties into the database
	for key, value := range properties {
		err := gen.DB.UpdatePropery(key, value)
		if err != nil {
			return err
		}
	}

	// disable any properties that are not in the server.properties file
	keys := "("
	for key := range properties {
		keys += fmt.Sprintf("'%s',", key)
	}
	keys = keys[:len(keys)-1] + ")"
	_, err = gen.DB.Exec(fmt.Sprintf("update server_properties set active = false where key not in %s", keys))
	if err != nil {
		return log.Error(err)
	}
	return nil
}

func readPlayerFile(file string) ([]map[string]interface{}, error) {
	// read the file and return a slice of jsonPlayers
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := io.Reader(f)
	decoder := json.NewDecoder(reader)
	players := []map[string]interface{}{}
	err = decoder.Decode(&players)
	if err != nil {
		return nil, err
	}
	return players, nil
}
