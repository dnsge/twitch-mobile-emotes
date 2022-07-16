# Build stage
FROM golang:1.18-alpine AS build

WORKDIR /go/src/github.com/dnsge/twitch-mobile-emotes

ADD go.mod .
ADD go.sum .
RUN go mod download

ADD . .
RUN go build -o /go/bin/github.com/dnsge/emote-server /go/src/github.com/dnsge/twitch-mobile-emotes/cmd/emote-server

# Final stage
FROM alpine

WORKDIR /app
COPY --from=build /go/bin/github.com/dnsge/emote-server /app/

ENTRYPOINT ["./emote-server"]
