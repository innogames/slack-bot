FROM golang:alpine AS builder

WORKDIR /code/
COPY . ./

RUN apk add git build-base
RUN go build -trimpath -ldflags="-s -w -X github.com/innogames/slack-bot/v2/bot/version.Version=$(git describe --tags 2>/dev/null || echo unknown)" -o /app cmd/bot/main.go

FROM alpine:latest AS alpine
RUN apk add --no-cache git ca-certificates tzdata
COPY --from=builder app .

CMD ["./app"]
