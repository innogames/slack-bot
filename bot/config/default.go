package config

var defaultConfig = Config{
	StoragePath: "./storage/",
	Logger: Logger{
		File:  "./bot.log",
		Level: "info",
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
