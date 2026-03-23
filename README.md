# Excur VK Bot — Setup & Deploy Guide

Bilingual (🇸🇦 Arabic / 🇬🇧 English) VK (VKontakte) currency exchange bot.
You set rates manually every day. Bot reminds you at noon, then every hour until you do.

Pairs: EGP↔RUB, USDT→RUB, USD→EGP, RUB→USDT  (AED→RUB ready to uncomment)

---

## How VK bots work (quick overview)

VK bots live inside a **VK Group (community)**, not a personal account.
Users message your group, and the bot replies. You need to:
1. Create a VK Group
2. Get an API token from that group
3. Enable messaging in the group settings

---

## Step 1 — Create a VK Group

1. Go to https://vk.com and log in
2. Click **"Create community"** (left sidebar or vk.com/groups)
3. Choose **"Business or brand"** or **"Public page"** — either works
4. Name it **Excur** (or whatever you like)
5. Note the group ID — it's the number in the URL: `vk.com/club123456789` → ID is `123456789`
   - This is your `VK_GROUP_ID`

---

## Step 2 — Enable messages in the group

1. Go to your group → **Manage** (top right)
2. Go to **Messages** section
3. Enable **"Community messages"** → Save

---

## Step 3 — Get a group API token

1. In your group → **Manage** → **API usage** (or "Working with API")
2. Click **"Create token"**
3. Enable these permissions:
   - ✅ Community messages
   - ✅ Community management (needed to send messages)
4. Confirm and copy the token
5. This is your `VK_GROUP_TOKEN`

---

## Step 4 — Enable Long Poll API

1. In your group → **Manage** → **API usage** → **Long Poll API** tab
2. Enable it → set version to **5.199**
3. Under **Event types**, enable:
   - ✅ Message received (incoming messages)

---

## Step 5 — Get your personal VK user ID

You need this so the bot knows where to send YOU the daily reminders.

1. Go to https://vk.com/id_getter  OR
2. Open your VK profile → the number in the URL is your ID
   e.g. `vk.com/id123456` → your ID is `123456`
3. This is your `ADMIN_VK_ID`

---

## Step 6 — Run locally (to test)

Requires Go: https://go.dev/dl/

```bash
cd excur-vk-bot

# Linux / Mac
export VK_GROUP_TOKEN="your_group_token"
export VK_GROUP_ID="123456789"
export ADMIN_VK_ID="your_personal_vk_id"

# Windows
set VK_GROUP_TOKEN=your_group_token
set VK_GROUP_ID=123456789
set ADMIN_VK_ID=your_personal_vk_id

go run .
```

You should see: `Excur VK bot started, listening for messages...`

Open VK, go to your group, click **Message** — you'll see the bot respond.

---

## Step 7 — Deploy online with Railway (free, easiest)

1. Push to GitHub:
   ```bash
   git init && git add . && git commit -m "excur vk bot"
   git remote add origin https://github.com/YOUR_NAME/excur-vk-bot.git
   git push -u origin main
   ```
2. Go to https://railway.app → sign up → New Project → Deploy from GitHub
3. Select your repo
4. Go to **Variables** tab and add:
   - `VK_GROUP_TOKEN` = your group token
   - `VK_GROUP_ID` = your group ID number
   - `ADMIN_VK_ID` = your personal VK user ID
5. Deploy — it runs 24/7!

---

## How the daily reminder works

- Every day at **12:00 PM** (Cairo time) the bot messages you listing missing rates
- If you don't set all rates, it reminds you **every hour** until you do
- If rates aren't set, the bot uses **yesterday's rates** and tells the customer
- If no rates exist at all, the bot says "rate unavailable"

To change the time or timezone, edit the top of `main.go`:
```go
const reminderHour = 12           // 0–23
const timezone     = "Africa/Cairo"
```

---

## How you (admin) set rates

Message your own group. As the admin you see the rate management panel.

- Tap a pair → type the rate → e.g. `89.50` means 1 EGP = 89.50 RUB
- ✅ = set today | 🕐 = using yesterday | ❌ = not set at all
- When all done → "✅ All rates set for today!"

Commands:
- `/rates` — open rate panel
- `/status` — see all rates at a glance

---

## Enabling AED → RUB

In `main.go`, uncomment this line in the `pairs` slice:

```go
// {"AED_RUB", "🇦🇪 AED → 🇷🇺 RUB", "🇦🇪 درهم ← 🇷🇺 روبل", "AED", "RUB"},
```

---

## File structure

```
excur-vk-bot/
├── main.go         ← all bot logic
├── go.mod          ← Go module file (no external dependencies!)
├── Dockerfile      ← for deployment
├── railway.toml    ← Railway config
├── rates.json      ← auto-created, stores today/yesterday rates
└── README.md       ← this file
```

Note: This VK bot uses **zero external Go packages** — only the standard library.
This makes it simpler to run and deploy than the Telegram version.
