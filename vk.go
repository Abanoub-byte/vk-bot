package main

// ─────────────────────────────────────────────────────────────────────────────
// vk.go — VK API client, message sending, and keyboard builders
// ─────────────────────────────────────────────────────────────────────────────

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// ── VK API types ──────────────────────────────────────────────────────────────

type VKResponse struct {
	Response json.RawMessage `json:"response"`
	Error    *VKError        `json:"error"`
}

type VKError struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_msg"`
}

type LongPollServer struct {
	Key    string `json:"key"`
	Server string `json:"server"`
	Ts     string `json:"ts"`
}

type LongPollResponse struct {
	Ts      string     `json:"ts"`
	Updates []LPUpdate `json:"updates"`
	Failed  int        `json:"failed"`
}

type LPUpdate struct {
	Type   string          `json:"type"`
	Object json.RawMessage `json:"object"`
}

type VKMessage struct {
	ID      int    `json:"id"`
	FromID  int    `json:"from_id"`
	Out     int    `json:"out"` // 1 = outgoing (bot's own message echo), skip these
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

type VKMessageObject struct {
	Message VKMessage `json:"message"`
}

// ── API call ──────────────────────────────────────────────────────────────────

func vkCall(method string, params map[string]string) (json.RawMessage, error) {
	vals := url.Values{}
	vals.Set("access_token", os.Getenv("VK_GROUP_TOKEN"))
	vals.Set("v", vkAPIVersion)
	for k, v := range params {
		vals.Set(k, v)
	}
	resp, err := http.PostForm(vkAPIBase+method, vals)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var vkResp VKResponse
	if err := json.Unmarshal(body, &vkResp); err != nil {
		return nil, err
	}
	if vkResp.Error != nil {
		return nil, fmt.Errorf("VK error %d: %s", vkResp.Error.Code, vkResp.Error.Message)
	}
	return vkResp.Response, nil
}

// sendMessage sends a text message to a VK user with an optional keyboard.
func sendMessage(userID int, text string, keyboard string) {
	log.Printf("→ user %d: %q", userID, text)
	params := map[string]string{
		"user_id":   strconv.Itoa(userID),
		"message":   text,
		"random_id": strconv.Itoa(rand.Int()),
	}
	if keyboard != "" {
		params["keyboard"] = keyboard
	}
	raw, err := vkCall("messages.send", params)
	if err != nil {
		log.Printf("sendMessage error to %d: %v", userID, err)
	} else {
		log.Printf("sendMessage ok to %d: %s", userID, string(raw))
	}
}

// ── Keyboard types ────────────────────────────────────────────────────────────

type VKKeyboard struct {
	OneTime bool         `json:"one_time"`
	Buttons [][]VKButton `json:"buttons"`
}

type VKButton struct {
	Action VKButtonAction `json:"action"`
	Color  string         `json:"color"`
}

type VKButtonAction struct {
	Type    string `json:"type"`
	Label   string `json:"label"`
	Payload string `json:"payload"`
}

// btn creates a secondary (grey) button.
func btn(label, payload string) VKButton {
	return VKButton{
		Color: "secondary",
		Action: VKButtonAction{
			Type:    "text",
			Label:   label,
			Payload: fmt.Sprintf(`{"cmd":%q}`, payload),
		},
	}
}

// primaryBtn creates a blue highlighted button.
func primaryBtn(label, payload string) VKButton {
	return VKButton{
		Color: "primary",
		Action: VKButtonAction{
			Type:    "text",
			Label:   label,
			Payload: fmt.Sprintf(`{"cmd":%q}`, payload),
		},
	}
}

// kb builds a keyboard JSON string from rows of buttons.
func kb(oneTime bool, rows ...[]VKButton) string {
	data, _ := json.Marshal(VKKeyboard{OneTime: oneTime, Buttons: rows})
	return string(data)
}

// ── Keyboard builders ─────────────────────────────────────────────────────────

func langKeyboard() string {
	return kb(true,
		[]VKButton{btn("English", "lang:en"), btn("Arabic", "lang:ar")},
	)
}

func pairKeyboard(lang Lang) string {
	t := tr(lang)
	var rows [][]VKButton
	for i, p := range pairs {
		label := p.LabelEN
		if lang == AR {
			label = p.LabelAR
		}
		rows = append(rows, []VKButton{primaryBtn(label, fmt.Sprintf("pair:%d", i))})
	}
	switchLabel, switchCmd := "Arabic", "lang:ar"
	if lang == AR {
		switchLabel, switchCmd = "English", "lang:en"
	}
	rows = append(rows, []VKButton{btn(t.About, "about"), btn(t.ContactBtn, "contact")})
	rows = append(rows, []VKButton{btn(switchLabel, switchCmd)})
	return kb(true, rows...)
}

func backKeyboard(lang Lang) string {
	return kb(true,
		[]VKButton{btn(tr(lang).Back, "menu")},
	)
}

func newConvKeyboard(lang Lang) string {
	t := tr(lang)
	return kb(true,
		[]VKButton{primaryBtn(t.NewConversion, "menu")},
		[]VKButton{btn(t.ContactBtn, "contact")},
	)
}

func adminKeyboard() string {
	var rows [][]VKButton
	for i, p := range pairs {
		_, today, ok := getRate(p.Key)
		status := "[X]"
		if ok && today {
			status = "[OK]"
		} else if ok && !today {
			status = "[~]"
		}
		rows = append(rows, []VKButton{
			btn(fmt.Sprintf("%s %s", status, p.LabelEN), fmt.Sprintf("admin_set:%d", i)),
		})
	}
	rows = append(rows, []VKButton{
		btn("View rates", "admin_view"),
		btn("Stats", "admin_stats"),
	})
	return kb(true, rows...)
}
