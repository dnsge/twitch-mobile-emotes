# twitch-mobile-emotes

An HTTP server that intercepts and modifies Twitch's IRC-over-WebSockets chat and its emoticon CDN to introduce BetterTTV and FrankerFaceZ custom emotes to native chats, specifically on mobile.

You must run this program behind some reverse proxy that can apply SSL certificates for the domains `static-cdn.jtvnw.net` and `irc-ws.chat.twitch.tv`. I choose to use a self-signed certificate that is trusted on my devices. See the `nginx` folder for an example.

## Building

Build from source
```bash
$ go build ./cmd/emote-server
```
...or use the [docker container](https://github.com/dnsge/twitch-mobile-emotes/packages/531933):
```bash
$ docker pull docker.pkg.github.com/dnsge/twitch-mobile-emotes/twitch-mobile-emotes:latest
```

## Usage

```
Usage of emote-server:
  -address string
    	Bind address (default "0.0.0.0:8080")
  -emoticon-host string
    	Host header to expect from Emoticon requests (default "emoticon.proxy")
  -no-gifs
    	Disable showing gif emotes
  -ws-host string
    	Host header to expect from Websocket IRC requests (default "irc-ws.proxy")
```

`emote-server` checks `Host` headers to determine which handler HTTP requests should be sent to (either emote CDN or IRC WS). 

Use the `--emoticon-host` and `--ws-host` flags to tell the server what to expect.

If you want to disable gif emotes, as they don't properly render on mobile, pass the `--no-gifs` flag.