package main

// ─────────────────────────────────────────────────────────────────────────────
// longpoll.go — VK long poll connection and update dispatch
// ─────────────────────────────────────────────────────────────────────────────

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func getLongPollServer() (*LongPollServer, error) {
	raw, err := vkCall("groups.getLongPollServer", map[string]string{
		"group_id": os.Getenv("VK_GROUP_ID"),
	})
	if err != nil {
		return nil, err
	}
	var lp LongPollServer
	if err := json.Unmarshal(raw, &lp); err != nil {
		return nil, err
	}
	return &lp, nil
}

// runLongPoll connects to VK's long poll server and dispatches incoming messages.
// This is the main blocking loop — it runs until the process exits.
func runLongPoll() {
	lp, err := getLongPollServer()
	if err != nil {
		log.Fatalf("getLongPollServer: %v", err)
	}
	log.Println("Excur VK bot started.")

	for {
		pollURL := fmt.Sprintf("%s?act=a_check&key=%s&ts=%s&wait=25",
			lp.Server, lp.Key, lp.Ts)

		resp, err := http.Get(pollURL)
		if err != nil {
			log.Printf("poll error: %v — retry in 5s", err)
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var lpResp LongPollResponse
		if err := json.Unmarshal(body, &lpResp); err != nil {
			log.Printf("unmarshal error: %v", err)
			continue
		}

		// Restart long poll on failure
		if lpResp.Failed != 0 {
			log.Printf("Long poll failed=%d, restarting", lpResp.Failed)
			lp, err = getLongPollServer()
			if err != nil {
				log.Printf("restart error: %v", err)
				time.Sleep(5 * time.Second)
			}
			continue
		}

		lp.Ts = lpResp.Ts

		for _, upd := range lpResp.Updates {
			if upd.Type != "message_new" {
				continue
			}
			var obj VKMessageObject
			if err := json.Unmarshal(upd.Object, &obj); err != nil {
				continue
			}
			msg := obj.Message
			if msg.FromID <= 0 || msg.Out == 1 {
				continue // skip system/outgoing echo messages
			}
			log.Printf("← user %d: text=%q payload=%q", msg.FromID, msg.Text, msg.Payload)
			handleMessage(msg.FromID, msg.Text, msg.Payload)
		}
	}
}
