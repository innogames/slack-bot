# GOWORK=off: this example uses a released slack-bot version and is
# intentionally not part of the repository's go.work workspace
build: export GOWORK=off
build:
	go mod tidy
	go build -o slack-bot-test *.go
