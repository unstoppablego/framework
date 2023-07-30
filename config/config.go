package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/unstoppablego/framework/logs"
)

// 尝试读取数据库信息
func Init() {

}

type Version struct {
	Version string
}

type DBConfig struct {
	User   string
	Passwd string
	Dbname string
	Host   string
	Port   string
	Other  string
	Tag    string
	Type   string
}

func (dc *DBConfig) DSN() string {
	return dc.User + ":" + dc.Passwd + "@tcp(" + dc.Host + ":" + "3306" + ")/" + dc.Dbname + "?" + dc.Other
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type HttpConfig struct {
	Port               string
	Address            string
	SetMaxIdleConns    int
	SetMaxOpenConns    int
	SetConnMaxLifetime int
	Doc                bool
}

type AliyunEmailConfig struct {
	AccessKeyID     string
	AccessKeySecret string
}

type ConfigV1 struct {
	Version
	DB          []DBConfig
	Aliyunemail AliyunEmailConfig
	Redis       []RedisConfig
	Http        HttpConfig
}

func ReadConf(Path string) {

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(Path)     // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	logs.Info("Read Config Sucess ")
	logs.Info(viper.GetString("db.0.user"))
	var ver Version
	viper.Unmarshal(&ver)
	logs.Info(ver.Version)

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

var Cfg ConfigV1
