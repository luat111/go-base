package main

type RedisOptions struct {
	CacheHost string `mapstructure:"CACHE_HOST"`
	CachePort string `mapstructure:"CACHE_PORT"`
	CachePass string `mapstructure:"CACHE_PWD"`
	CacheDB   int    `mapstructure:"CACHE_DB"`
}

type AppConfig struct {
	//Environment
	AppPort string `mapstructure:"PORT"`
	RpcPort string `mapstructure:"RPC_PORT"`
	AppName string `mapstructure:"APP_NAME"`
	ENV     string `mapstructure:"ENV" json:"ENV"`

	//Base Route
	API_PATH string `mapstructure:"API_PATH"`

	//Redis
	CacheOptions RedisOptions `mapstructure:"CACHE"`
}
