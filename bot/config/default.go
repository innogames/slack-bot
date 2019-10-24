package config

var defaultConfig = Config{
	StoragePath: "./storage/",
	Logger: Logger{
		File: "./bot.log",
	},
	Server: Server{
		Listen: ":8765",
	},

	Timezone: "Europe/Berlin",

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
