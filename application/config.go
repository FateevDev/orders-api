package application

import "os"

type Config struct {
	RedisAddress string
	Port         string
}

func LoadConfig() Config {
	cfg := Config{
		RedisAddress: "localhost:6380",
		Port:         "3000",
	}

	if redisAdd, exists := os.LookupEnv("REDIS_ADDRESS"); exists {
		cfg.RedisAddress = redisAdd
	}

	if port, exists := os.LookupEnv("PORT"); exists {
		cfg.Port = port
	}

	return cfg
}
