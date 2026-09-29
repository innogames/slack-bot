# builds an image with a bot binary including all plugins of the config.yaml
FROM golang:alpine AS builder
RUN apk add --no-cache git

WORKDIR /build
COPY . .

ARG BOT_VERSION=latest
# for private Go modules, e.g. "docker build --build-arg GOPRIVATE=gitlab.example.com"
ARG GOPRIVATE
RUN go run github.com/innogames/slack-bot/v2/cmd/slack-bot-builder@${BOT_VERSION} -config config.yaml -output /slack-bot

FROM alpine:latest
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY --from=builder /slack-bot ./slack-bot
COPY config.yaml ./config.yaml

CMD ["./slack-bot", "-config", "config.yaml"]
