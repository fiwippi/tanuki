package logging

type Config struct {
	Level        Level `yaml:"level"`
	LogToFile    bool  `yaml:"log_to_file"`
	LogToConsole bool  `yaml:"log_to_console"`
}

func DefaultConfig() *Config {
	return &Config{
		Level:        InfoLevel,
		LogToFile:    true,
		LogToConsole: true,
	}
}
