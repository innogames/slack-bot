package config

// DefaultConfig with some common values
var DefaultConfig = Config{
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
	PullRequest: PullRequest{
		Reactions: PullRequestReactions{
			InReview:     "eyes",
			Approved:     "white_check_mark",
			Merged:       "twisted_rightwards_arrows",
			Closed:       "x",
			BuildFailed:  "fire",
			BuildRunning: "arrows_counterclockwise",
			Error:        "x",
		},
	},
}
