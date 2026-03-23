package main

// ─────────────────────────────────────────────────────────────────────────────
// handlers.go — all message routing and business logic
// ─────────────────────────────────────────────────────────────────────────────

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

// ── Admin helpers ─────────────────────────────────────────────────────────────

func adminID() int {
	id, _ := strconv.Atoi(os.Getenv("ADMIN_VK_ID"))
	return id
}

func isAdmin(userID int) bool {
	return userID == adminID()
}

func sendAdminMenu(userID int) {
	sessions[userID] = &UserState{Step: StepPickPair}
	sendMessage(userID,
		"Rate Management\n[OK]=today  [~]=yesterday  [X]=not set",
		adminKeyboard(),
	)
}

// handleAdminRate processes a rate value typed by the admin.
func handleAdminRate(userID int, text string, state *UserState) {
	log.Printf("handleAdminRate pairIdx=%d text=%q", state.AdminPairIdx, text)

	rate, err := strconv.ParseFloat(
		strings.TrimSpace(strings.ReplaceAll(text, ",", ".")), 64)
	if err != nil || rate <= 0 {
		sendMessage(userID, "Invalid. Enter a number like: 89.50", "")
		return
	}

	p := pairs[state.AdminPairIdx]
	rolloverIfNeeded()
	store.Rates[p.Key] = rate
	saveRates()
	log.Printf("Saved: %s = %.4f (%d total)", p.Key, rate, len(store.Rates))

	suffix := ""
	if todayRatesComplete() {
		suffix = "\n\nAll rates set for today!"
	}
	sessions[userID] = &UserState{Step: StepPickPair}
	sendMessage(userID,
		fmt.Sprintf("Saved: %s\n1 %s = %.4f %s%s", p.LabelEN, p.From, rate, p.To, suffix),
		adminKeyboard(),
	)
}

// ── Main message router ───────────────────────────────────────────────────────

func handleMessage(userID int, text, payload string) {
	// Parse button command from payload JSON
	cmd := ""
	if payload != "" {
		var p map[string]string
		if err := json.Unmarshal([]byte(payload), &p); err == nil {
			cmd = p["cmd"]
		}
	}

	state := getState(userID)
	trackUser(userID, state.Lang)

	// Admin: waiting for rate input (must check before converting text to cmd)
	if isAdmin(userID) && state.Step == StepAdminSetRate && cmd == "" {
		handleAdminRate(userID, text, state)
		return
	}

	// User: waiting for contact info
	if state.Step == StepContact && cmd == "" {
		handleContactSubmit(userID, text, state)
		return
	}

	// No payload — treat raw text as a command
	if cmd == "" {
		cmd = strings.TrimSpace(strings.ToLower(text))
	}

	if cmd != "" {
		trackButton(cmd)
	}

	// User: waiting for amount to convert
	if state.Step == StepEnterAmount && cmd == "" {
		handleAmount(userID, text, state)
		return
	}

	// Route to the right handler
	switch {
	case cmd == "/start" || cmd == "start" || cmd == "/menu" || cmd == "menu":
		handleStart(userID)

	case cmd == "lang:en" || cmd == "lang:ar":
		handleLangSelect(userID, cmd, state)

	case strings.HasPrefix(cmd, "pair:"):
		handlePairSelect(userID, cmd, state)

	case cmd == "about":
		handleAbout(userID, state)

	case cmd == "contact":
		handleContactStart(userID, state)

	case cmd == "/rates" && isAdmin(userID):
		sendAdminMenu(userID)

	case cmd == "/status" && isAdmin(userID):
		sendMessage(userID, currentRatesSummary(), adminKeyboard())

	case cmd == "/stats" && isAdmin(userID):
		sendMessage(userID, statsSummary(), adminKeyboard())

	case strings.HasPrefix(cmd, "admin_set:") && isAdmin(userID):
		handleAdminSetStart(userID, cmd)

	case cmd == "admin_view" && isAdmin(userID):
		sendMessage(userID, currentRatesSummary(), adminKeyboard())

	case cmd == "admin_stats" && isAdmin(userID):
		sendMessage(userID, statsSummary(), adminKeyboard())

	case cmd == "admin_menu" && isAdmin(userID):
		sendAdminMenu(userID)

	default:
		// Fallback: if user was mid-conversion, try as amount
		if state.Step == StepEnterAmount {
			handleAmount(userID, text, state)
		} else if isAdmin(userID) {
			sendAdminMenu(userID)
		} else {
			handleStart(userID)
		}
	}
}

// ── Individual handlers ───────────────────────────────────────────────────────

func handleStart(userID int) {
	if isAdmin(userID) {
		sendAdminMenu(userID)
		return
	}
	sessions[userID] = &UserState{Step: StepPickLang, Lang: EN}
	sendMessage(userID, "Choose language / اختر اللغة:", langKeyboard())
}

func handleLangSelect(userID int, cmd string, state *UserState) {
	lang := EN
	if cmd == "lang:ar" {
		lang = AR
	}
	state.Lang = lang
	state.Step = StepPickPair
	sessions[userID] = state
	sendMessage(userID, tr(lang).Welcome, pairKeyboard(lang))
}

func handlePairSelect(userID int, cmd string, state *UserState) {
	idx, _ := strconv.Atoi(strings.TrimPrefix(cmd, "pair:"))
	if idx < 0 || idx >= len(pairs) {
		return
	}
	p := pairs[idx]
	lang := state.Lang

	sessions[userID] = &UserState{
		Step:     StepEnterAmount,
		Lang:     lang,
		FromCurr: p.From,
		ToCurr:   p.To,
		PairKey:  p.Key,
		LabelEN:  p.LabelEN,
		LabelAR:  p.LabelAR,
	}

	var prompt string
	if lang == AR {
		prompt = fmt.Sprintf(tr(lang).YouChose, p.LabelEN)
	} else {
		prompt = fmt.Sprintf(tr(lang).YouChose, p.LabelEN, p.From)
	}
	sendMessage(userID, prompt, backKeyboard(lang))
}

func handleAbout(userID int, state *UserState) {
	lang := state.Lang
	t := tr(lang)
	sendMessage(userID, t.AboutText,
		kb(true,
			[]VKButton{btn(t.ContactBtn, "contact")},
			[]VKButton{btn(t.Back, "menu")},
		),
	)
}

func handleContactStart(userID int, state *UserState) {
	lang := state.Lang
	sessions[userID] = &UserState{Step: StepContact, Lang: lang}
	sendMessage(userID, tr(lang).ContactPrompt, backKeyboard(lang))
}

func handleContactSubmit(userID int, text string, state *UserState) {
	lang := state.Lang
	t := tr(lang)

	trackContact(userID, text)

	// Notify admin with the user's VK profile link
	sendMessage(adminID(),
		fmt.Sprintf("New contact request!\nvk.com/id%d\n\n%s", userID, text),
		adminKeyboard(),
	)

	sessions[userID] = &UserState{Step: StepPickPair, Lang: lang}
	sendMessage(userID, t.ContactThanks, pairKeyboard(lang))
}

func handleAdminSetStart(userID int, cmd string) {
	idx, _ := strconv.Atoi(strings.TrimPrefix(cmd, "admin_set:"))
	if idx < 0 || idx >= len(pairs) {
		return
	}
	p := pairs[idx]
	sessions[userID] = &UserState{Step: StepAdminSetRate, AdminPairIdx: idx}
	sendMessage(userID,
		fmt.Sprintf("Enter rate for %s\nExample: 89.50\n(1 %s = X %s)", p.LabelEN, p.From, p.To),
		kb(true, []VKButton{btn("Cancel", "admin_menu")}),
	)
}

// handleAmount processes a conversion amount from a regular user.
func handleAmount(userID int, text string, state *UserState) {
	lang := state.Lang
	t := tr(lang)

	amount, err := strconv.ParseFloat(
		strings.TrimSpace(strings.ReplaceAll(text, ",", ".")), 64)
	if err != nil || amount <= 0 {
		sendMessage(userID, t.InvalidAmount, backKeyboard(lang))
		return
	}

	rate, isToday, ok := getRate(state.PairKey)
	if !ok {
		sendMessage(userID, t.RateUnavailable, backKeyboard(lang))
		return
	}

	excurRate := rate * (1 + markupPercent/100)
	result := math.Round(amount*excurRate*100) / 100

	yesterdayNote := ""
	if !isToday {
		yesterdayNote = "\n" + t.UsingYesterday
	}

	var reply string
	if lang == AR {
		// Arabic result: short English-only format to avoid RTL/LTR rendering issues
		reply = fmt.Sprintf(
			"RESULT: %.2f %s => %.2f %s  RATE: %.2f%s",
			amount, state.FromCurr,
			result, state.ToCurr,
			excurRate,
			yesterdayNote,
		)
	} else {
		reply = fmt.Sprintf(
			"%s\n\n%s: %s\n%s: %.2f %s\n%s: 1 %s = %.2f %s\n\n%s: %.2f %s%s",
			t.ResultTitle,
			t.Pair, state.PairLabel(),
			t.Amount, amount, state.FromCurr,
			t.YourRate, state.FromCurr, excurRate, state.ToCurr,
			t.YouReceive, result, state.ToCurr,
			yesterdayNote,
		)
	}

	sessions[userID] = &UserState{Step: StepPickPair, Lang: lang}
	sendMessage(userID, reply, newConvKeyboard(lang))
}
