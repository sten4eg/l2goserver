package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var globalConfig Conf

type Conf struct {
	LoginServer LoginServerType `yaml:"loginserver"`
	GameServer  GameServerType  `yaml:"gameserver"`
}

type DatabaseType struct {
	Name        string `yaml:"name"`
	Host        string `yaml:"host"`
	Schema      string `yaml:"schema"`
	Port        string `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	SSLMode     string `yaml:"sslMode"`
	PoolMaxConn string `yaml:"poolMaxConn"`
}

type LoginServerType struct {
	Host                 string       `yaml:"host"`
	AutoCreate           bool         `yaml:"autoCreate"`
	PortForGS            string       `yaml:"portForGS"`
	Database             DatabaseType `yaml:"database"`
	AllowedServerVersion []byte       `yaml:"allowedServerVersion"`
}

type GameServerType struct {
	Name       string   `yaml:"name"`
	InternalIp string   `yaml:"internalIp"`
	Port       string   `yaml:"port"`
	MaxPlayers uint16   `yaml:"maxPlayers"`
	HexIds     [][]byte `yaml:"hexIds"`
}

func Read() error {
	var config Conf
	file, err := os.Open("./config/config.yaml")
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(file)
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

func GetGameServerHexId() [][]byte {
	return globalConfig.GameServer.HexIds
}
