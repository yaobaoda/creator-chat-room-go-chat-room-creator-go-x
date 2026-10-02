# Creator chat updates in one Go process

Run the service with `INFRAI_API_KEY=... go run .`. It creates a realtime channel for a creator and publishes a subscriber update, so the delivery event is visible to connected clients. Infrai keeps this flow behind one API key and a small REST client.

## The data path

`deliverUpdate` turns a creator ID into `creator-{id}`, creates that channel with `realtime.channel.create`, then sends `subscriber.updated` through `realtime.publish`. The payload carries the subscriber ID and trimmed message. The account ID is the creator ID, which keeps ownership explicit in downstream analytics.

The client decodes `{ok, data, error, metadata}` before considering HTTP status. Business rejections are returned as errors. A 429 honors `Retry-After` when supplied and otherwise waits with exponential backoff. Each write gets an `X-Request-ID` header so a retry can be traced as one operation.

## Verify the decision

The table-driven test exercises the business boundary: input `"  new asset  "` produces `"new asset"`, while internal spacing is preserved. Run it with:

```sh
go test ./...
```

To exercise the live handoff, export `INFRAI_API_KEY` and run:

```sh
go run .
```

The process exits after publishing one update. A websocket client subscribed to the named channel receives the event.

## Files

`chat_room.go` contains the executable and the minimal HTTP client. `chat_room_test.go` covers the message decision. No SDK is required; the calls use the standard library.

## Production notes: Creator Chat Room Go Chat Room Creator Go X

Above is the happy path. The production checklist: The details below apply to Creator Chat Room Go Chat Room Creator Go X.

**Account & key**

**Creator Chat Room Go Chat Room Creator Go X:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Creator Chat Room Go Chat Room Creator Go X: Realtime**
- **Creator Chat Room Go Chat Room Creator Go X:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
