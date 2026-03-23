package main

// ─────────────────────────────────────────────────────────────────────────────
// config.go — app constants, currency pairs, and language definitions
// ─────────────────────────────────────────────────────────────────────────────

const (
	markupPercent = 0.0             // % added on top of manual rates (0 = disabled)
	reminderHour  = 12              // hour to send daily rate reminder (24h clock)
	timezone      = "Africa/Cairo"  // your local timezone
	vkAPIVersion  = "5.199"
	vkAPIBase     = "https://api.vk.com/method/"
	ratesFile     = "rates.json"
	statsFile     = "stats.json"
)

// ── Currency pairs ────────────────────────────────────────────────────────────

type PairKey string

var pairs = []struct {
	Key     PairKey
	LabelEN string
	LabelAR string
	From    string
	To      string
}{
	{"EGP_RUB", "EGP → RUB", "EGP → RUB", "EGP", "RUB"},
	{"RUB_EGP", "RUB → EGP", "RUB → EGP", "RUB", "EGP"},
	{"AED_RUB", "AED → RUB", "AED → RUB", "AED", "RUB"},
	{"RUB_AED", "RUB → AED", "RUB → AED", "RUB", "AED"},
	// {"AED_EGP", "AED → EGP", "AED → EGP", "AED", "EGP"}, // uncomment to add
}



// ── Languages & translations ──────────────────────────────────────────────────

type Lang string

const (
	EN Lang = "en"
	AR Lang = "ar"
)

type Strings struct {
	Welcome         string
	PickPair        string
	YouChose        string
	InvalidAmount   string
	RateUnavailable string
	ResultTitle     string
	Pair            string
	Amount          string
	YourRate        string
	YouReceive      string
	UsingYesterday  string
	NewConversion   string
	Back            string
	About           string
	AboutText       string
	ChooseLang      string
	ContactBtn      string
	ContactPrompt   string
	ContactThanks   string
	UpdateRates     string
}

var translations = map[Lang]Strings{
	EN: {
		Welcome:         "Welcome to Excur - choose a currency to know the exchange rate:\n",
		PickPair:        "Choose a currency pair:",
		YouChose:        "Pair: %s\n\nEnter amount in %s:",
		InvalidAmount:   "Invalid number. Example: 500",
		RateUnavailable: "Rate not available yet. Try again soon.",
		ResultTitle:     "Excur - Result",
		Pair:            "Pair",
		Amount:          "Amount",
		YourRate:        "Rate",
		YouReceive:      "You receive",
		UsingYesterday:  "(yesterday rate)",
		NewConversion:   "New conversion",
		Back:            "Back",
		About:           "About Excur",
		AboutText:       "Excur - testing bot for vk.\nFor knowing rates.\n",
		ChooseLang:      "Choose language:",
		ContactBtn:      "Contact owner",
		ContactPrompt:   "Send your phone number and Telegram username.\n\nExample:\n+7 999 123 4567\n@myusername",
		ContactThanks:   "Thank you! We will contact you soon.",
		UpdateRates:     "Update rates",
	},
	AR: {
		Welcome:         "اهلا في اكسكور لمعرفة اسعار العملات العملات. اختر زوج العملات:",
		PickPair:        "اختر زوج العملات:",
		YouChose:        "PAIR: %s  ادخل المبلغ:",
		InvalidAmount:   "رقم غير صحيح. مثال: 500",
		RateUnavailable: "السعر غير متاح. حاول لاحقا.",
		ResultTitle:     "RESULT",
		Pair:            "PAIR",
		Amount:          "AMT",
		YourRate:        "RATE",
		YouReceive:      "ستستلم",
		UsingYesterday:  "(سعر امس)",
		NewConversion:   "تحويل جديد",
		Back:            "رجوع",
		About:           "عن التطبيق",
		AboutText:       "اكسكور - خدمة لمعرفة اسعار العملات.",
		ChooseLang:      "اختر اللغة:",
		ContactBtn:      "تواصل مع المالك",
		ContactPrompt:   "ارسل رقم هاتفك وحساب تيليغرام. مثال: +7 999 123 4567 @myusername",
		ContactThanks:   "شكرا! سنتواصل معك قريبا.",
		UpdateRates:     "تحديث الاسعار",
	},
}

func tr(lang Lang) Strings { return translations[lang] }
