package main

// ─────────────────────────────────────────────────────────────────────────────
// main.go — entry point
// ─────────────────────────────────────────────────────────────────────────────

import (
	"log"
	"os"
)

func main() {
	// Validate required environment variables
	if os.Getenv("VK_GROUP_TOKEN") == "" {
		log.Fatal("VK_GROUP_TOKEN env var is required")
	}
	if os.Getenv("VK_GROUP_ID") == "" {
		log.Fatal("VK_GROUP_ID env var is required")
	}
	if adminID() == 0 {
		log.Fatal("ADMIN_VK_ID env var is required (your numeric VK user ID)")
	}

	// Initialise data
	loadRates()
	loadStats()
	rolloverIfNeeded()

	// Start background scheduler (daily/hourly rate reminders)
	startScheduler()

	// Start the main long poll loop (blocking)
	runLongPoll()
}
