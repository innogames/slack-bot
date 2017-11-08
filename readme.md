# Slack Bot
This slack bot improves the workflow of development teams. Especially with focus on Jenkins and Jira integration.

[![Build Status](https://travis-ci.org/innogames/slack-bot.svg)](https://travis-ci.org/innogames/slack-bot)
[![GoDoc](https://godoc.org/github.com/innogames/slack-bot?status.svg)](https://godoc.org/github.com/innogames/slack-bot)
[![Go Report Card](https://goreportcard.com/badge/github.com/innogames/slack-bot)](https://goreportcard.com/report/github.com/innogames/slack-bot)
[![Release](https://img.shields.io/github/release/innogames/slack-bot.svg)](https://github.com/innogames/slack-bot/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

# Usage
As slack user, you just have to send a private message to the bot user/app containing the command to execute.
Additionally you can execute bot commands in channels by prefix your command with @bot_name, e.g. `@slack-bot start job DailyDeployment`

**Note:** You have to invite the bot into the channel to be able to handle commands.

# Commands
## Help
The `help` command just prints a list of all available commands of this bot. 
With `help *command*` you'll get a short description and some examples for a single command.

## Jenkins
The bot is able to start and monitor jenkins job on a simple but powerful way.

### Start Jenkins jobs
The `start job` command starts a Jenkins job and shows the current progress. **Attention:** only whitelisted jobs in the config are startable!

In additions each job can have a configurable `trigger` which make it possible to create custom commands to start jobs. (it's a regexp which takes parameter names into account).
E.g. "start daily deployment" could be the trigger for one jenkins job. Sending this text to the bot would start he job.

After starting a job the bot will show the estimated build time and some action buttons. There you can open the logs or abort the build directly.

The bot is also able to parse parameters and lookup branch names using a fuzzy branch search.

**Examples:**
- `trigger job DeployBeta`
- `start job BackendTests TEST-123` (search for a full branch name, containing TEST-123. e.g. feature/TEST-123-added-feature-456)

![Screenshot](./docs/jenkins-trigger.png)

### Jenkins build notifications
The bot has also the possibility to create one time notifications for jenkins builds. This might be useful for long running jobs where the devs is waiting for the result.

**Example:**
- `inform me about build NightlyTests` (watches the most recent running build)
- `inform me about build MyJobName #423` (specify a build number)
- `inform job MyJobName` (alternative syntax)

### Jenkins job notifications
Receive slack messages for all process builds for the given job:
**Example:**
- `watch JenkinsMonitoring`
- `unwatch Jenkins Monitoring`

### Jenkins status
Small command to disable/enable job execution on Jenkins side.
**Example:**
- `disable job NightlyTests` (disable job on jenkins)
- `enable job NightlyTests`

### Jenkins retry
When a build failed you are able to retry any build by:
**Example:**
- `retry build NightlyTests` (retries the last build of a job)
- `retry build NightlyTests #100` (retries given build)

## Pull Requests
If you just paste a link to a Github/Gitlab/Stash Pull request, the bot will track the state of the ticket! 
- When a developer was added as reviewer, it will add a "eyes" reaction to show other devs that someone is already taking a look
- When the reviewer approved the ticket, a checkmark is added
- After merging the pull request, it will add a "merge" reaction

![Screenshot](./docs/pull-request.png)

## Command Queue
The `queue` command (with the alias `then`) is able to queue the given command, until the currently running command finished. 

Example following scenario: you have a build job (which might take some minutes) and a deploy job which relies of the build artifacts. Now you can do:
- `trigger job Build feature1234` to start the Build job with given branch
- `queue trigger job DeployBranch feature1234` 
- `queue reply Deployment is done!`

**Other example:**
- `delay 1h`
- `then send message #backend coffee time?`

To see all running background commands (like Jenkins jobs or PR watcher) use this command:
- `list queue`

## Jira
Query information from Jira, either from a single ticket, or a whole list of tickets.

**Examples**
- `jira TEST-1234`
- `jira 1242` (opens the ticket, using the configured default jira project)
- `jql type=bug and status=open` (use default project by default)
- `jira "Second city"` (text search of tickets in default project)

## Custom command
Every user is able to define own command aliases. This is a handy feature to avoid tying the same command every day.

**Commands**
- `list commands`
- `add command 'myCommand' 'trigger job RestoreWorld 7` -> then just call `myCommand` later
- `add command 'build master' 'trigger job Deploy master ; then trigger job DeployClient master'`
- `delete command 'build master'`
- -> then you can execute `myCommand` to trigger this jenkins job
![Screenshot](./docs/custom-commands.png)

## Macro
Macros are very magical and can be defined in the "config.yaml". 
They have a trigger (a regular expression) and have a list of sub commands which will be executed. They take parameter groups from regexp into account - so they can be very flexible!

One simple example to start two Jenkins jobs with a given branch name at the same time:
```
macros:
 - name: build clients
   trigger: "build clients (?P<branch>.*)"
   commands:
    - "reply I'll build {{ .branch }} for you"
    - "trigger job BuildFrontendClient {{ .branch }}"
    - "trigger job BuildMobileClient {{ .branch }}"
```
![Screenshot](./docs/macro-multiple-jobs.jpg)

**Note**: In the commands you can use the full set of [template features of go](https://golang.org/pkg/text/template/) -> loops/conditions are possible!

## Retry
With `retry` or `repeat` your last executed command will be re-executed. -> Useful when a failed Jenkins job got fixed.

## Delay
A small command which might be useful in combination with `macro` command or as hook for jenkins jobs.

Example command: `delay 10m trigger job DeployWorldwide`

As reply you'll get a command to stop the queued job (like `stop timer 123456`). As everyone can send the command, the command can be used to announce a deployment and in doubt, the execution can still be stopped by everyone.

## Reply / send message
`reply` and `send message` are also small commands which are useful in combination with `macro` or jenkins hooks.

**Examples:**
- `send message to #backend The job failed :panic:`
- `delay 10m send message to @peter_pan I should notify you to...`

## Random
Simple command if you are not able to decide between different options

**Examples**
- `random Pizza Pasta` -> produce either "Pizza" or "Pasta" 
- `random Peter Paul Tom Jan` -> who has to take about organizing food today?

# Installation
1. clone the project
2. create config file called `config.yaml` (you can take a look at `config.example.yaml`)
 
## Run without docker
This command will start the bot, using the `config.yaml` file by default. Use the `-config` argument to use the config file(s) from another location.
```
go run cmd/bot/main.go
```

## Run via docker-compose
**Attention**: Create a config.yaml file first

```
docker-compose up --build
```

# Configuration
The configuration is managed via simple yaml files which are storing the credentials for the external services and the custom commands etc.
It's supported to split up the configuration into multiple files.
**Possible structure:**
- `secret.yaml` containing the credentials for the external services (slack, jenkins) - can be managed by puppet/ansible etc.
- `jenkins.yaml` configuration of jenkins job and their parameters etc
- `project-X.yaml` custom commands (aka macros) for a specific team
- `project-Y.yaml`

To load the config files, use `go run cmd/bot/main.go -config /path/to/config/*.yaml` which merged all configs together.

## Slack
To run this bot, you need a "bot token" for your slack application. [Take a look here](https://api.slack.com/docs/token-types#bot) how to get one.

You can define a "team" field: This will grant access to all users which are allocated to this team name. As alternative, the "allowedusers" option can be used to grant access. (will change in future)
```
slack:
  token: xoxb-1234567-secret
  team: "Development"
```

## Jenkins
To be able to start or monitor jenkins jobs, you have to setup the host and the credentials first. The user needs read access to the jobs and the right to trigger jobs for your whitelisted jobs.
```
jenkins:
     host: https://jenkins.example.de
     username: jenkinsuser
     password: secret
```

### Jenkins jobs
To be able to start a job, the job and it's parameters have to be defined in the config.

A job without any parameter looks very simple:
```
jenkins:
  jobs:
    CleanupJob:
```
Then you can use `trigger job CleanupJob` or `start job CleanupJob` to start the job. It will also notify you when the job succeeded or failed (incl. error log). 

Next a job with two parameters:
```
jenkins:
  jobs:
    RunTests:
      parameters:
      - name: BRANCH
        default: master
        type: branch
      - name: GROUP
        default: all
```
This job can handle two parameters:
 - BRANCH: VCS branch name, "master" as default
 - GROUP: optional parameter, using "all" as default
        
If you setup the VSC in the config, you don't have to pass the full branch name but can use the fuzzy search.

**Example:**
 - `start job RunTests` would start "all" groups on master branch
 - `start job JIRA-1224 unit` would try to find a matching branch for the ticket number. (Error message if there is no unique search result!)
        
Now a more complex example with more magic: 
```
jenkins:
     jobs:
       DeployBranch:
         trigger: "deploy (?P<BRANCH>[\\w\\-_\\.\\/]*) to (?P<ENVIRONMENT>prod|test|env)"
         parameters:
         - name: BRANCH
           default: master
           type: branch
         - name: ENVIRONMENT
         onsuccess:
          - reply Tadaa: Take a look on http://{{ .ENVIRONMENT }}.example.com
```
**Step by step:**
The `trigger` is a regular expression to start the job which may contain named groups. The regexp groups will be matched to the job parameters automatically.

Then you can use `deploy bugfix-1234 to test` to start the jenkins job.

**Note:** You can always start this job also via `start job DeployBranch master`. The `trigger` is just an alternative.

The `onsuccess` is a hook which will be executed when a job ist started via this bot. 
In addition `onsuccess` and `onerror` is also available...e.g. to send custom error messages.

### MQTT
MQTT is a simple publish-subscribe messaging protocol, based on TCP/IP.
**Example config**
```
mqtt:
  host: tcp://localhost:1883
``` 

**Commands**
```
- mqtt subscribe temperature
- mqtt publish temperature 1.22
- mqtt unsubscribe temperature
```

### Cron
**Example config**
```
crons:
  - schedule: "0 8 * * *"
    commands:
      - trigger job BuildClients
      - then deploy master to staging
    channel: "#backend"
```

### Calendar
Trigger commands by calendar entries of an ical/icl calenar.
**Example:**
```
calendars:
  - path: https://calendar.google.com/calendar/ical/ic2sdfafdsfdsfdsfsdfds5d0c19f8/basic.ics
    events:
    - name: Create release branch
      trigger: "Create release branch (?P<branch>.*)"
      commands:
      - "trigger job CreateReleaseBranch {.branch}"
      - "send message to #release I'll created branch {.branch}"
    - name: "beer time"
      trigger: "beer time"
      commands:
      - "send message to #all :beer:"

```
The calendar appointment "Create release branch 2.124" will start the given jenkins job and post a message into #release channel

## VCS / Stash / Bitbucket
To be able to resolve branch names in jenkins trigger, a VCS system can be configured (at the moment it's just Stash/Bitbucket).
```
vcs:
  type: bitbucket
  host: https://bitbucket.example.com
  username: readonlyuser
  password: secret
  project: MyProjectKey
  repository: repo_name
```
If no config is provided, there is no automated branch lookup and the "branch" parameters are passed 1:1 to the jenkins job.

# Development

## File structure
- `bot` contains the code classes of the bot: connection to slack, user management, command matching
- `cmd` entry points aka main.go for the bot and the CLI tool
- `command` real command implementations impleenting the bot.Command interace

## Create a new (native) command
If you need a new command, which is not implementable with a "macro" command, you have to write to write go code.
- create a new file within the "commands/" directory or one submodule of it
- create a new struct which fulfills the bot.Command interface. The service.SlackClient might be needed as dependency
- GetMatcher() needs to provide the information which command text is matching our command
- register the command in command/commands.go
- restart the bot application
- it's recommended to fulfill the bot.HelpProvider (your command will show up in `help)
- it's also recommended to create a integration test for your command

## CLI tool
There is a handy CLI application which emulates the Slack application...just chat with your local console without any Slack connection!
```
go run cmd/cli/main.go
Type in your command:
delay 1s reply IT WORKS!
>>>> delay 1s reply IT WORKS!
<<<< I queued the command `reply IT WORKS!` for 1s. Use `stop timer 0` to stop the timer
<<<< IT WORKS!
```

## Testing
There are a bunch of tests which can be executed via:
```
make test
```

Test coverage is generated to build/coverage.html
```
make test-coverage
```

## Benchmarks
```
make test-bench
```
