package main

// ─────────────────────────────────────────────────────────────────────────────
// scheduler.go — daily rate reminders and hourly follow-ups
// ─────────────────────────────────────────────────────────────────────────────

import (
	"fmt"
	"log"
	"time"
)

// startScheduler runs the daily/hourly reminder loop in a background goroutine.
func startScheduler() {
	go func() {
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			log.Printf("Bad timezone %q, falling back to UTC", timezone)
			loc = time.UTC
		}

		var lastReminderDay string
		lastHourlyHour := -1

		for {
			now := time.Now().In(loc)
			today := now.Format("2006-01-02")
			hour := now.Hour()

			rolloverIfNeeded()

			// Daily reminder at reminderHour
			if hour == reminderHour && lastReminderDay != today {
				lastReminderDay = today
				lastHourlyHour = -1
				if !todayRatesComplete() {
					sendReminderToAdmin(false)
				}
			}

			// Hourly follow-up if rates still not set
			if hour > reminderHour && lastHourlyHour != hour && !todayRatesComplete() {
				lastHourlyHour = hour
				sendReminderToAdmin(true)
			}

			// Sleep until the next minute boundary
			next := now.Truncate(time.Minute).Add(time.Minute)
			time.Sleep(time.Until(next))
		}
	}()
}

func sendReminderToAdmin(isHourly bool) {
	id := adminID()
	if id == 0 {
		log.Println("ADMIN_VK_ID not set — skipping reminder")
		return
	}
	missing := missingSummary()
	if missing == "" {
		return
	}
	prefix := "Daily reminder - set today's exchange rates!"
	if isHourly {
		prefix = "Reminder - rates not set yet. Using yesterday's rates."
	}
	sendMessage(id, fmt.Sprintf("%s\n\nMissing:\n%s", prefix, missing), adminKeyboard())
}
