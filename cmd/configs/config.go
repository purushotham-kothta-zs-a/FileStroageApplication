package configs

import "os"

type Config struct {
	FileStoragePath string
}

func LoadConfig() *Config {
	return &Config{
		os.Getenv("FILE_STORAGE_PATH"),
	}
}
