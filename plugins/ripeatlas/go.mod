module github.com/innogames/slack-bot/plugins/ripeatlas

go 1.26.0

require (
	github.com/innogames/slack-bot/v2 v2.3.18
	github.com/sirupsen/logrus v1.9.4
	github.com/slack-go/slack v0.27.0
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bndr/gojenkins v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gfleury/go-bitbucket-v1 v0.0.0-20240917142304-df385efaac68 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/hackebrot/turtle v0.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.0 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/redis/go-redis/v9 v9.21.0 // indirect
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/texttheater/golang-levenshtein/levenshtein v0.0.0-20200805054039-cae8b0eaed6c // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20260709172345-9ea1abe57597 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// use the slack-bot of this repository for the development and tests of the plugin.
// It's ignored when the plugin is used as dependency, e.g. by the slack-bot-builder
replace github.com/innogames/slack-bot/v2 => ../..
