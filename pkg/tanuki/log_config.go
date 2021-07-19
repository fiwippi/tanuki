package tanuki

//
type logConfig struct {
	Level        LogLevel `yaml:"level"`
	LogToFile    bool     `yaml:"log_to_file"`
	LogToConsole bool     `yaml:"log_to_console"`
}

//
func defaultLogConfig() logConfig {
	return logConfig{
		Level:     InfoLevel,
		LogToFile:    true,
		LogToConsole: true,
	}
}
