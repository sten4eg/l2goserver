package config

import (
	"encoding/json"
	"os"
)

var globalConfig Conf

type Conf struct {
	LoginServer LoginServerType `json:"loginserver"`
	GameServer  GameServerType  `json:"gameserver"`
}

type DatabaseType struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	SSLMode     string `json:"sslMode"`
	PoolMaxConn string `json:"PoolMaxConn"`
}

type LoginServerType struct {
	Host                 string       `json:"host"`
	AutoCreate           bool         `json:"autoCreate"`
	PortForGS            string       `json:"portForGS"`
	Database             DatabaseType `json:"database"`
	AllowedServerVersion []byte       `json:"allowedServerVersion"`
}

type GameServerType struct {
	Name       string `json:"name"`
	InternalIp string `json:"internalIp"`
	Port       string `json:"port"`
	MaxPlayers uint16 `json:"maxPlayers"`
	HexId      []byte `json:"hexId"`
}

func Read() error {
	var config Conf
	file, err := os.Open("./config/config.json")
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	globalConfig = config
	return nil
}

func AutoCreateAccounts() bool {
	return globalConfig.LoginServer.AutoCreate
}

func GetLoginPortForGameServer() string {
	return globalConfig.LoginServer.PortForGS
}
func GetConfig() Conf {
	return globalConfig
}

func GetAllowedServerVersion() []byte {
	return globalConfig.LoginServer.AllowedServerVersion
}
func GetGameServerHexId() []byte {
	return globalConfig.GameServer.HexId
}
