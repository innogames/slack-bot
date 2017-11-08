package config

var defaultConfig = Config{
	StoragePath: "./storage/",
	Logger: Logger{
		File: "./bot.log",
	},
}
