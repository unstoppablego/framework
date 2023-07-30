package test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/unstoppablego/framework/app"
	"github.com/unstoppablego/framework/config"
)

func TestDB(t *testing.T) {
	flag.Int("flagname", 1234, "help message for flagname")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	i := viper.GetInt("flagname") // retrieve value from viper

	app.Log.Info(i)
	viper.SetConfigName("config")        // name of config file (without extension)
	viper.SetConfigType("yaml")          // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/appname/") // path to look for the config file in
	viper.AddConfigPath("$HOME/config/") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	viper.AddConfigPath("../config/")    // optionally look for config in the working directory
	err := viper.ReadInConfig()          // Find and read the config file
	if err != nil {                      // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	app.Log.Info("Read Config Sucess ")
	app.Log.Info(viper.GetString("db.0.user"))
	var ver config.Version
	viper.Unmarshal(&ver)
	app.Log.Info(ver.Version)
	var c1 config.ConfigV1
	viper.Unmarshal(&c1)
	app.Log.Info(c1.Version, c1.DB[0].Dbname)
}
