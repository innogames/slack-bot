package config

var defaultConfig = Config{
	StoragePath: "./storage/",
	Logger: Logger{
		File:  "./bot.log",
		Level: "info",
	},
	Server: Server{
		Listen: ":8765",
	},
	OpenWeather: OpenWeather{
		Units: "metric",
	},
	// some default jira fields
	Jira: Jira{
		Fields: []JiraField{
			{
				Name: "type",
				Icons: map[string]string{
					"Bug": ":bug:",
				},
			},
		},
	},
}
