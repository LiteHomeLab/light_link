package client

import (
    "encoding/json"

    "github.com/nats-io/nats.go"
)

// MessageHandler message handler
type MessageHandler func(data map[string]interface{})

// Subscription represents a subscription
type Subscription struct {
    sub *nats.Subscription
}

// Unsubscribe unsubscribes
func (s *Subscription) Unsubscribe() error {
    if s.sub != nil {
        return s.sub.Unsubscribe()
    }
    return nil
}

// Publish publishes a message
func (c *Client) Publish(subject string, data map[string]interface{}) error {
    msgData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    return c.nc.Publish(subject, msgData)
}

// Subscribe subscribes to messages
func (c *Client) Subscribe(subject string, handler MessageHandler) (*Subscription, error) {
    sub, err := c.nc.Subscribe(subject, func(msg *nats.Msg) {
        var data map[string]interface{}
        if err := json.Unmarshal(msg.Data, &data); err != nil {
            return
        }
        handler(data)
    })

    return &Subscription{sub: sub}, err
}
