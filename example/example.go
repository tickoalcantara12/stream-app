package main

import (
	"fmt"
	"time"

	sdk "github.com/tickoalcantara12/stream-app"
)

func main() {
	cfg := sdk.DefaultConfig()
	cfg.Endpoint = "ws://localhost:8080/ws"
	cfg.HeartbeatInterval = 5 * time.Second

	c := sdk.NewClient(cfg)
	if err := c.Connect(); err != nil {
		panic(err)
	}
	defer func(c *sdk.Client) {
		err := c.Close()
		if err != nil {
			return
		}
	}(c)

	c.OnEvent(func(e sdk.Event) {
		fmt.Println("EVENT:", e.Type, e.Data)
	})

	streamID, err := c.StartStream(sdk.Metadata{Title: "Demo Stream"})
	if err != nil {
		panic(err)
	}
	fmt.Println("Started stream:", streamID)

	_ = c.SendVideoEvent(streamID, map[string]string{"viewer": "user-1"})
	time.Sleep(1 * time.Second)

	_ = c.EndStream(streamID)
}
