package main

// ─────────────────────────────────────────────────────────────────────────────
// stats.go — user analytics, button tracking, contact request log
// ─────────────────────────────────────────────────────────────────────────────

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type UserInfo struct {
	ID        int    `json:"id"`
	FirstSeen string `json:"first_seen"`
	LastSeen  string `json:"last_seen"`
	Lang      string `json:"lang"`
	MsgCount  int    `json:"msg_count"`
}

type ContactEntry struct {
	UserID    int    `json:"user_id"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
}

type StatsStore struct {
	Users        map[string]*UserInfo `json:"users"`
	ButtonCounts map[string]int       `json:"button_counts"`
	Contacts     []ContactEntry       `json:"contacts"`
}

var stats StatsStore

func loadStats() {
	data, err := os.ReadFile(statsFile)
	if err != nil {
		stats = StatsStore{
			Users:        map[string]*UserInfo{},
			ButtonCounts: map[string]int{},
			Contacts:     []ContactEntry{},
		}
		return
	}
	if err := json.Unmarshal(data, &stats); err != nil {
		stats = StatsStore{
			Users:        map[string]*UserInfo{},
			ButtonCounts: map[string]int{},
			Contacts:     []ContactEntry{},
		}
	}
	if stats.Users == nil {
		stats.Users = map[string]*UserInfo{}
	}
	if stats.ButtonCounts == nil {
		stats.ButtonCounts = map[string]int{}
	}
	if stats.Contacts == nil {
		stats.Contacts = []ContactEntry{}
	}
}

func saveStats() {
	data, _ := json.MarshalIndent(stats, "", "  ")
	os.WriteFile(statsFile, data, 0644)
}

func trackUser(userID int, lang Lang) {
	key := strconv.Itoa(userID)
	now := time.Now().Format("2006-01-02 15:04")
	if u, ok := stats.Users[key]; ok {
		u.LastSeen = now
		u.MsgCount++
		u.Lang = string(lang)
	} else {
		stats.Users[key] = &UserInfo{
			ID:        userID,
			FirstSeen: now,
			LastSeen:  now,
			Lang:      string(lang),
			MsgCount:  1,
		}
	}
	saveStats()
}

func trackButton(label string) {
	stats.ButtonCounts[label]++
	saveStats()
}

func trackContact(userID int, text string) {
	stats.Contacts = append(stats.Contacts, ContactEntry{
		UserID:    userID,
		Text:      text,
		Timestamp: time.Now().Format("2006-01-02 15:04"),
	})
	saveStats()
}

// statsSummary returns a formatted analytics report for the admin.
func statsSummary() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Stats\n\nTotal users: %d\n\n", len(stats.Users)))

	// Top buttons sorted by count
	type kv struct {
		Key string
		Val int
	}
	var sorted []kv
	for k, v := range stats.ButtonCounts {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Val > sorted[j].Val })
	sb.WriteString("Top buttons:\n")
	for i, kv := range sorted {
		if i >= 8 {
			break
		}
		sb.WriteString(fmt.Sprintf("  %s: %d\n", kv.Key, kv.Val))
	}

	// Most recent 10 users
	type userKV struct {
		id   string
		info *UserInfo
	}
	var users []userKV
	for k, v := range stats.Users {
		users = append(users, userKV{k, v})
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].info.LastSeen > users[j].info.LastSeen
	})
	sb.WriteString("\nRecent users:\n")
	for i, u := range users {
		if i >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("  vk.com/id%d | %s | %s | %d msgs\n",
			u.info.ID, u.info.Lang, u.info.LastSeen, u.info.MsgCount))
	}

	// Most recent 5 contact requests
	if len(stats.Contacts) > 0 {
		sb.WriteString("\nContact requests:\n")
		start := len(stats.Contacts) - 5
		if start < 0 {
			start = 0
		}
		for _, c := range stats.Contacts[start:] {
			sb.WriteString(fmt.Sprintf("  vk.com/id%d [%s]: %s\n",
				c.UserID, c.Timestamp, c.Text))
		}
	}
	return sb.String()
}
