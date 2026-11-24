FROM golang:alpine AS builder

WORKDIR /code/
COPY . ./

RUN apk add git build-base
# retry the build to be robust against transient errors from the Go module proxy
RUN version=$(git describe --tags 2>/dev/null || echo unknown); \
    for i in 1 2 3; do \
        go build -trimpath -ldflags="-s -w -X github.com/innogames/slack-bot/v2/bot/version.Version=$version" -o /app cmd/bot/main.go && break; \
        [ "$i" -eq 3 ] && exit 1; \
        echo "go build failed, retrying ($i/3)"; sleep $((i * 10)); \
    done

FROM alpine:latest AS alpine
RUN apk add --no-cache git ca-certificates tzdata
COPY --from=builder app .

CMD ["./app"]
