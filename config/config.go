package config

import (
	"log"
	"os"
	"strconv"
)

// database .env for app
type DatabaseConfig struct {
	DB_TYPE string
	DB_HOST string
	DB_PORT string
	DB_NAME string
	DB_USER string
	DB_PASS string
	OTHER_P map[string]string
}

// accumulate env
type Config struct {
	DBConfig          DatabaseConfig
	CurrentAppVersion string
	Debug_mode        bool
	Hostname          string
	TCPPort           string
}

// singleton instance
var config *Config

// init config module
func GetConfig() *Config {
	if config == nil {

		other_p := make(map[string]string)
		other_p["DB_SSLMODE"] = getEnv("DB_SSLMODE")
		other_p["DB_TZ"] = getEnv("DB_TZ")

		config = &Config{
			DBConfig: DatabaseConfig{
				DB_TYPE: getEnv("DB_TYPE"),
				DB_HOST: getEnv("DB_HOST"),
				DB_PORT: getEnv("DB_PORT"),
				DB_NAME: getEnv("DB_NAME"),
				DB_USER: getEnv("DB_USER"),
				DB_PASS: getEnv("DB_PASS"),
				OTHER_P: other_p,
			},
			CurrentAppVersion: getEnv("APP_VERSION"),
			Debug_mode:        getBoolEnv("DEBUG_MODE"),
			Hostname:          getEnv("HOSTNAME"),
			TCPPort:           getEnv("PORT"),
		}
	}
	return config
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func getFloatEnv(key string) float32 {
	val, err := strconv.ParseFloat(os.Getenv(key), 32)
	if err != nil {
		log.Fatalf("error while parse value: %s", err.Error())
	}
	return float32(val)
}

func getIntEnv(key string) int32 {
	val, err := strconv.ParseInt(os.Getenv(key), 10, 32)
	if err != nil {
		log.Fatalf("error while parse value: %s", err.Error())
	}
	return int32(val)
}

func getBoolEnv(key string) bool {
	val, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		log.Fatalf("error while parse value: %s", err.Error())
	}
	return val
}

func getUIntEnv(key string) uint32 {
	val, err := strconv.ParseUint(os.Getenv(key), 10, 32)
	if err != nil {
		log.Fatalf("error while parse value: %s", err.Error())
	}
	return uint32(val)
}
