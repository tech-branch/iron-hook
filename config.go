package ironhook

import "github.com/spf13/viper"

func initConfig() {

	viper.SetEnvPrefix("hook")
	viper.AutomaticEnv()

	// defaults
	//
	// database
	viper.SetDefault("db_engine", "sqlite")
	viper.SetDefault("db_dsn", ":memory:")
	//
	// logger
	viper.SetDefault("log_level", "info")

}
