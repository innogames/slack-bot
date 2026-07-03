module github.com/innogames/slack-bot/v2/plugins/example

go 1.25.0

// within this repository the slack-bot dependency is resolved via the go.work workspace;
// external users need a slack-bot release which includes the plugin API (> v2.3.17)
require (
	github.com/innogames/slack-bot/v2 v2.3.17
	github.com/stretchr/testify v1.11.1
)
