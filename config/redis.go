package config

// Redis defines the config of redis.
type Redis struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	DB       int    `config:"db"`
	Username string `config:"username"`
	Password string `config:"password"`
}
