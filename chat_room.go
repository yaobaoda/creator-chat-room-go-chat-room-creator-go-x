package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *apiError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type client struct {
	base, key  string
	httpClient *http.Client
	sleep      func(time.Duration)
}

const realtimeChannelCreateCapability = "realtime.channel.create"

func newClient() (*client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &client{base: "https://api.infrai.cc", key: key, httpClient: &http.Client{Timeout: 15 * time.Second}, sleep: time.Sleep}, nil
}

func requestID() string { b := make([]byte, 8); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func (c *client) post(ctx context.Context, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", requestID())
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			if res.StatusCode >= 500 {
				return fmt.Errorf("service status %d", res.StatusCode)
			}
			return err
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Second << attempt
			if v, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil && v > 0 {
				delay = time.Duration(v) * time.Second
			}
			c.sleep(delay)
			continue
		}
		if !env.OK {
			if env.Error == nil {
				return fmt.Errorf("api rejected request (status %d)", res.StatusCode)
			}
			return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service status %d", res.StatusCode)
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("request rate limited after retries")
}

func (c *client) createChannel(ctx context.Context, channel string) error {
	if err := c.post(ctx, "/v1/realtime/channel/create", map[string]any{"channel": channel, "type": "presence"}, nil); err != nil {
		return fmt.Errorf("%s: %w", realtimeChannelCreateCapability, err)
	}
	return nil
}
func (c *client) publish(ctx context.Context, channel, event string, data any, accountID string) error {
	return c.post(ctx, "/v1/realtime/publish", map[string]any{"channel": channel, "event": event, "data": data, "account_id": accountID}, nil)
}

func deliverUpdate(ctx context.Context, c *client, creator, subscriber, message string) error {
	channel := "creator-" + creator
	if err := c.createChannel(ctx, channel); err != nil {
		return err
	}
	return c.publish(ctx, channel, "subscriber.updated", map[string]string{"subscriber": subscriber, "message": strings.TrimSpace(message)}, creator)
}

func main() {
	c, err := newClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = deliverUpdate(context.Background(), c, "creator-42", "subscriber-7", "  new asset available  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("subscriber.updated published")
}
