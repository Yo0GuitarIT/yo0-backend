# yo0-backend

Go 後端服務，整合 Unsplash、中央氣象署（CWA）、Telegram Bot 與 Discord Bot，提供照片查詢、天氣摘要、塔羅占卜與推播功能。

## 技術棧

- **Go 1.25**
- **Gin** — HTTP 框架
- **Air** — Hot reload 工具
- **Docker Compose** — 開發環境
- **Telegram Bot API** — 訊息互動與推播
- **Discord Bot (discordgo)** — 按鈕互動式塔羅占卜
- **CWA Open Data API** — 24 小時天氣資料

## 專案結構

```
yo0-backend/
├── cmd/
│   └── main.go                 # 程式入口
├── internal/
│   ├── handler/
│   │   ├── notify.go           # 測試推播 API
│   │   ├── photo.go            # 照片 API
│   │   ├── tarot.go            # 塔羅牌 API
│   │   └── weather.go          # 天氣 API
│   ├── model/
│   │   ├── tarot.go            # Tarot 資料結構
│   │   ├── photo.go            # Photo 資料結構
│   │   └── weather.go          # Weather 資料結構
│   ├── service/
│   │   ├── telegram_bot.go     # Bot 啟動與指令派發（Long Polling）
│   │   ├── telegram_commands.go# 各指令 handler（/weather、/image、/tarot 等）
│   │   ├── telegram_formatter.go# 訊息格式化工具
│   │   ├── telegram_push.go    # 早安推播邏輯（天氣 + 照片 + 塔羅）
│   │   ├── telegram_scheduler.go# 定時排程（每日 06:00 台灣時間）
│   │   ├── telegram_store.go   # 用戶預設城市記憶體儲存
│   │   ├── discord_bot.go      # Discord Bot 啟動與按鈕互動監聽
│   │   ├── discord_tarot.go    # Discord 塔羅占卜互動邏輯
│   │   ├── tarot.go            # 塔羅牌抽牌邏輯（78 張完整牌組）
│   │   ├── tarot_image.go      # 塔羅牌圖片處理（逆位旋轉）
│   │   ├── unsplash.go         # Unsplash API 邏輯
│   │   └── weather.go          # CWA 天氣資料清洗
│   └── router/
│       └── router.go           # 路由註冊
├── .env                        # 環境變數（不 commit）
├── .air.toml                   # Hot reload 設定
├── docker-compose.yml
└── Dockerfile.dev
```

## 環境設定

建立 `.env` 檔案：

```
UNSPLASH_ACCESS_KEY=你的_Access_Key
TELEGRAM_BOT_TOKEN=你的_Bot_Token
TELEGRAM_CHAT_ID=你的_Chat_ID
CWB_API_KEY=你的_CWA_Authorization
DISCORD_BOT_TOKEN=你的_Discord_Bot_Token
DISCORD_CHANNEL_ID=你的_Discord_頻道_ID
```

> 到 [unsplash.com/oauth/applications](https://unsplash.com/oauth/applications) 取得 Access Key。
>
> 到 [中央氣象署開放資料平台](https://opendata.cwa.gov.tw/) 取得授權碼（Authorization）。
>
> 到 [Discord Developer Portal](https://discord.com/developers/applications) 建立 Bot 並取得 Token；`DISCORD_CHANNEL_ID` 為要貼出抽牌按鈕的文字頻道 ID。

備註：天氣金鑰同時支援 `CWB_API_KEY` 或 `CWA_API_KEY`。Discord 相關金鑰為選填，未設定時 Discord Bot 會自動略過。

## 啟動開發環境

本地只需要安裝 **Docker Desktop**，不需要安裝 Go。

```bash
# 啟動（含 hot reload）
docker compose up

# 停止
docker compose down

# 改了 Dockerfile.dev 才需要加 --build
docker compose up --build
```

## API

| Method | Path               | 說明                                             |
| ------ | ------------------ | ------------------------------------------------ |
| GET    | `/`                | 健康檢查                                         |
| GET    | `/ping`            | 測試 Telegram 連線與發送訊息                     |
| GET    | `/photos/random`   | 取得一張隨機照片                                 |
| GET    | `/weather/current` | 取得指定城市 24 小時重點天氣（預設臺南市）       |
| GET    | `/tarot/random`    | 隨機抽一張塔羅牌（78 張完整牌組，含正逆位）      |
| POST   | `/notify/test`     | 立即觸發一次早安推播（天氣 + 隨機照片 + 塔羅牌） |

### 範例

```bash
# 健康檢查
curl http://localhost:8080/

# 隨機照片
curl http://localhost:8080/photos/random

# 天氣（預設城市：臺南市）
curl "http://localhost:8080/weather/current"

# 天氣（指定城市）
curl "http://localhost:8080/weather/current?locationName=高雄市"

# 隨機塔羅牌
curl "http://localhost:8080/tarot/random"

# 測試推播（使用 .env 的 TELEGRAM_CHAT_ID）
curl -X POST "http://localhost:8080/notify/test"

# 測試推播（指定 chatId）
curl -X POST "http://localhost:8080/notify/test?chatId=5291868928"
```

## Telegram Bot 指令

| 指令              | 說明                                      |
| ----------------- | ----------------------------------------- |
| `/menu`           | 顯示所有可用指令                          |
| `/weather`        | 查詢個人預設城市 24 小時天氣              |
| `/weather 城市名` | 查詢指定城市天氣（例：`/weather 高雄市`） |
| `/setcity 城市名` | 設定個人預設城市（例：`/setcity 臺南市`） |
| `/mycity`         | 查看目前設定的預設城市                    |
| `/image`          | 取得一張 Unsplash 隨機照片                |
| `/tarot`          | 抽一張塔羅牌（含正逆位與牌面圖片）        |

## Discord Bot

Discord Bot 使用按鈕互動式占卜，不需要文字指令。

**互動流程：**

1. Bot 在指定頻道發送帶有「🔮 抽牌」按鈕的訊息
2. 使用者點擊按鈕後，Bot 回傳塔羅牌結果（牌名 + 牌面圖片，逆位自動旋轉 180°）
3. 被點過的按鈕自動變灰（不可重複點擊），同時在結果下方補出新的抽牌按鈕

**技術細節：**

- 使用 Deferred Response 處理逆位圖片下載與旋轉的延遲，避免 Discord 3 秒逾時
- 僅需 `Guilds` Intent，無需開啟 Message Content 特權意圖
- 未設定 `DISCORD_BOT_TOKEN` 時，服務啟動會自動略過 Discord Bot

## 早安推播

每天台灣時間 **06:00** 自動發送至 Telegram，內容依序為：

1. ☀️ 個人預設城市當日 24 小時天氣摘要
2. 🖼 一張 Unsplash 隨機照片
3. 🔮 每日塔羅牌（隨機抽取，逆位時牌面自動旋轉 180°）

也可透過 `POST /notify/test` 隨時手動觸發。

備註：預設城市存在記憶體（in-memory），服務重啟後會回到預設值 `臺南市`。

## 安裝新套件

因為本地沒有安裝 Go，透過容器執行：

```bash
docker compose run --rm app go get <套件名稱>
```
