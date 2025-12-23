package client

import (
    "context"
    "encoding/json"

    "github.com/nats-io/nats.go/jetstream"
)

// SetState sets state
func (c *Client) SetState(key string, value map[string]interface{}) error {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return err
    }

    // Get or create KV bucket
    kv, err := js.KeyValue(context.Background(), "light_link_state")
    if err != nil {
        // Create bucket
        kv, err = js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
            Bucket: "light_link_state",
        })
        if err != nil {
            return err
        }
    }

    // Serialize value
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    // Set KV
    _, err = kv.Put(context.Background(), key, data)
    return err
}

// GetState gets state
func (c *Client) GetState(key string) (map[string]interface{}, error) {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return nil, err
    }

    kv, err := js.KeyValue(context.Background(), "light_link_state")
    if err != nil {
        return nil, err
    }

    entry, err := kv.Get(context.Background(), key)
    if err != nil {
        return nil, err
    }

    var value map[string]interface{}
    if err := json.Unmarshal(entry.Value(), &value); err != nil {
        return nil, err
    }

    return value, nil
}

// WatchState watches state changes
func (c *Client) WatchState(key string, handler func(map[string]interface{})) (func(), error) {
    js, err := jetstream.New(c.nc)
    if err != nil {
        return nil, err
    }

    kv, err := js.KeyValue(context.Background(), "light_link_state")
    if err != nil {
        return nil, err
    }

    watcher, err := kv.Watch(context.Background(), key, jetstream.IgnoreDeletes())
    if err != nil {
        return nil, err
    }

    stop := make(chan struct{})

    go func() {
        for {
            select {
            case <-stop:
                watcher.Stop()
                return
            default:
                select {
                case entry := <-watcher.Updates():
                    if entry != nil {
                        var value map[string]interface{}
                        json.Unmarshal(entry.Value(), &value)
                        handler(value)
                    }
                case <-stop:
                    watcher.Stop()
                    return
                }
            }
        }
    }()

    return func() { close(stop) }, nil
}
