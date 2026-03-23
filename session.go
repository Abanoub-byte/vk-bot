package main

// ─────────────────────────────────────────────────────────────────────────────
// session.go — per-user conversation state (in-memory)
// ─────────────────────────────────────────────────────────────────────────────

type Step int

const (
	StepPickLang    Step = iota // waiting for user to choose language
	StepPickPair                // waiting for user to choose a currency pair
	StepEnterAmount             // waiting for user to type an amount
	StepAdminSetRate            // admin: waiting to type a rate value
	StepContact                 // user: waiting to type contact info
)

type UserState struct {
	Step         Step
	Lang         Lang
	FromCurr     string
	ToCurr       string
	PairKey      PairKey
	LabelEN      string
	LabelAR      string
	AdminPairIdx int // which pair the admin is currently setting
}

// PairLabel returns the pair label in the user's chosen language.
func (s *UserState) PairLabel() string {
	if s.Lang == AR {
		return s.LabelAR
	}
	return s.LabelEN
}

// sessions holds the current state for every active user, keyed by VK user ID.
var sessions = map[int]*UserState{}

// getState returns the current state for a user, creating a default if none exists.
func getState(userID int) *UserState {
	if s, ok := sessions[userID]; ok {
		return s
	}
	s := &UserState{Step: StepPickLang, Lang: EN}
	sessions[userID] = s
	return s
}
