module github.com/innogames/slack-bot/v2

go 1.15

require (
	github.com/alicebob/miniredis/v2 v2.15.1
	github.com/andygrunwald/go-jira v1.13.0
	github.com/bndr/gojenkins v1.1.0
	github.com/gfleury/go-bitbucket-v1 v0.0.0-20210707202713-7d616f7c18ac
	github.com/go-redis/redis/v7 v7.4.1
	github.com/google/go-github v17.0.0+incompatible
	github.com/gookit/color v1.4.2
	github.com/hackebrot/turtle v0.2.0
	github.com/jcelliott/lumber v0.0.0-20160324203708-dd349441af25 // indirect
	github.com/nanobox-io/golang-scribble v0.0.0-20190309225732-aa3e7c118975
	github.com/pkg/errors v0.9.1
	github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/slack-go/slack v0.9.3
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/texttheater/golang-levenshtein/levenshtein v0.0.0-20200805054039-cae8b0eaed6c
	github.com/xanzy/go-gitlab v0.50.1
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/text v0.3.6
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/spf13/viper => github.com/brainexe/viper v0.0.0-20201112092033-7bf4d99562ca
