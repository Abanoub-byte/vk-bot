package main

// ─────────────────────────────────────────────────────────────────────────────
// rates.go — daily rate storage, persistence, and rollover logic
// ─────────────────────────────────────────────────────────────────────────────

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type DayRates map[PairKey]float64

type RateStore struct {
	Today     string   `json:"today"`
	Rates     DayRates `json:"rates"`
	Yesterday DayRates `json:"yesterday"`
}

var store RateStore

func loadRates() {
	data, err := os.ReadFile(ratesFile)
	if err != nil {
		store = RateStore{Rates: DayRates{}, Yesterday: DayRates{}}
		return
	}
	if err := json.Unmarshal(data, &store); err != nil {
		store = RateStore{Rates: DayRates{}, Yesterday: DayRates{}}
	}
	if store.Rates == nil {
		store.Rates = DayRates{}
	}
	if store.Yesterday == nil {
		store.Yesterday = DayRates{}
	}
}

func saveRates() {
	data, _ := json.MarshalIndent(store, "", "  ")
	os.WriteFile(ratesFile, data, 0644)
}

func todayStr() string {
	loc, _ := time.LoadLocation(timezone)
	return time.Now().In(loc).Format("2006-01-02")
}

// rolloverIfNeeded moves today's rates to yesterday when the date changes.
func rolloverIfNeeded() {
	today := todayStr()
	if store.Today != today {
		if len(store.Rates) > 0 {
			store.Yesterday = store.Rates
		}
		store.Rates = DayRates{}
		store.Today = today
		saveRates()
	}
}

// getRate returns the best available rate for a pair.
// Prefers today's rate, falls back to yesterday's.
func getRate(key PairKey) (rate float64, isToday bool, ok bool) {
	rolloverIfNeeded()
	if r, exists := store.Rates[key]; exists && r > 0 {
		return r, true, true
	}
	if r, exists := store.Yesterday[key]; exists && r > 0 {
		return r, false, true
	}
	return 0, false, false
}

// todayRatesComplete returns true if all pairs have a rate set for today.
func todayRatesComplete() bool {
	rolloverIfNeeded()
	for _, p := range pairs {
		if _, ok := store.Rates[p.Key]; !ok {
			return false
		}
	}
	return true
}

// missingSummary returns a newline-separated list of pairs missing today's rate.
func missingSummary() string {
	var missing []string
	for _, p := range pairs {
		if _, ok := store.Rates[p.Key]; !ok {
			missing = append(missing, p.LabelEN)
		}
	}
	return strings.Join(missing, "\n")
}

// currentRatesSummary returns a human-readable summary of all current rates.
func currentRatesSummary() string {
	rolloverIfNeeded()
	var sb strings.Builder
	sb.WriteString("Current rates for today:\n\n")
	for _, p := range pairs {
		rate, today, ok := getRate(p.Key)
		if !ok {
			sb.WriteString(fmt.Sprintf("- %s: not set\n", p.LabelEN))
		} else if !today {
			sb.WriteString(fmt.Sprintf("- %s: %.4f (yesterday)\n", p.LabelEN, rate))
		} else {
			sb.WriteString(fmt.Sprintf("- %s: %.4f\n", p.LabelEN, rate))
		}
	}
	return sb.String()
}
