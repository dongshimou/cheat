package model

type ConfigDatabase struct {
	Host string `json:"host"`
	Port string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	MaxConn int `json:"max_conn"`
	MaxIdle int `json:"max_idle"`
}

type Config struct {
	Database ConfigDatabase `json:"database"`
}
