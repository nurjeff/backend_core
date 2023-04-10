package initbundle

import "gorm.io/gorm/logger"

type InitConfiguration struct {
	GinMode            string        `json:"gin_mode"`
	DBLoggerConfig     logger.Config `json:"db_logger_config"`
	ConfigPath         string        `json:"config_path"`
	MaxMultipartMemory int64         `json:"max_multipart_memory"`
}

type Config struct {
	AutoMigrate      bool      `json:"auto_migrate"`
	Database         Database  `json:"database"`
	Cache            Cache     `json:"cache"`
	Server           Server    `json:"server"`
	LogServer        LogServer `json:"logserver"`
	Mongo            Mongo     `json:"mongo"`
	Salt             string    `json:"salt"`
	JWTSecret        string    `json:"jwt_secret"`
	JWTRefreshSecret string    `json:"jwt_refresh_secret"`
}

type Mongo struct {
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Database struct {
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Server struct {
	Host string `json:"host"`
	Port uint   `json:"port"`
}

type LogServer struct {
	ProjectName string `json:"project_name"`
	RemoteLogs  bool   `json:"remote_logs"`
	Host        string `json:"host"`
	Port        uint   `json:"port"`
	ProjectKey  string `json:"project_key"`
	Client      string `json:"client"`
	ClientKey   string `json:"client_key"`
}

type Cache struct {
	CacheEngine  string `json:"cache_engine"`
	Address      string `json:"address"`
	PortOverride uint   `json:"port_override"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Workspace    string `json:"workspace"`
}
