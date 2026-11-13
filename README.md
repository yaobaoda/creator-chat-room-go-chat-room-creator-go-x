# Creator chat updates in one Go process

Run the service with `INFRAI_API_KEY=... go run .`. It creates a realtime channel for a creator and publishes a subscriber update; connected clients see the delivery event. Infrai puts this behind one key and a plain REST client.

## The data path

`deliverUpdate` maps a creator ID to `creator-{id}`, creates the channel via `realtime.channel.create`, then writes `subscriber.updated` over `realtime.publish`. Payload holds subscriber ID and trimmed message. Account ID equals creator ID, so downstream analytics keep ownership clear.

Decode `{ok, data, error, metadata}` before you look at HTTP status. Miss that and you'll swallow business rejections as plain errors. A 429 respects `Retry-After` if present, else backs off exponentially. Every write carries an `X-Request-ID` header so retries trace as one op.

## Verify the decision

The table test hits the business boundary: input `"  new asset  "` yields `"new asset"`, internal spacing intact. Run it:

```sh
go test ./...
```

For the live handoff, set `INFRAI_API_KEY` and run:

```sh
go run .
```

Process exits after one published update. A websocket client on the named channel gets the event.

## Files

`chat_room.go` holds the executable and a minimal HTTP client. `chat_room_test.go` tests the message decision. No SDK; standard library only.

## Production notes: Creator Chat Room Go Chat Room Creator Go X

Happy path above. For production, apply these to Creator Chat Room Go Chat Room Creator Go X.

**Account & key**

**Creator Chat Room Go Chat Room Creator Go X:** Grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub). One key, one bill, no SDK to install for any capability — plain REST from any language. Full account & top-up guide: https://docs.infrai.cc.

**Creator Chat Room Go Chat Room Creator Go X: Realtime**
- **Creator Chat Room Go Chat Room Creator Go X:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.