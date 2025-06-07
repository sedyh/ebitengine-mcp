package event

import (
	"encoding/json"
	"fmt"
	u "net/url"
	"time"

	"github.com/jcuga/golongpoll/client"
)

type Client[T Event] struct {
	client *client.Client
}

func NewClient[T Event](url, pub, sub, id string) (*Client[T], error) {
	var v T
	c, err := client.NewClient(client.ClientOptions{
		SubscribeUrl:   u.URL{Scheme: "http", Host: url, Path: sub},
		PublishUrl:     u.URL{Scheme: "http", Host: url, Path: pub},
		Category:       Mark(id, v.Type()),
		OnFailure:      retry(500 * time.Millisecond),
		LoggingEnabled: false,
	})
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	return &Client[T]{client: c}, nil
}

func (c *Client[T]) Start(since time.Time) <-chan T {
	ch := make(chan T)
	go func() {
		var v T
		for event := range c.client.Start(since) {
			data, err := json.Marshal(event.Data)
			if err != nil {
				v.SetError(fmt.Errorf("marshal json to bytes: %w", err))
				ch <- v
				continue
			}
			if err := json.Unmarshal(data, &v); err != nil {
				v.SetError(fmt.Errorf("unmarshal bytes to type: %w", err))
				ch <- v
				continue
			}
			ch <- v
		}
	}()
	return ch
}

func (c *Client[T]) Stop() {
	c.client.Stop()
}

func (c *Client[T]) Publish(id string, data T) error {
	return c.client.Publish(Mark(id, data.Type()), data)
}

func retry(delay time.Duration) func(error) bool {
	retries := 0
	return func(err error) bool {
		retries += 1
		if retries <= 3 {
			<-time.After(delay)
			return true
		}
		return false
	}
}
