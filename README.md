# yo0-backend

Go 後端服務，整合 Unsplash、中央氣象署（CWA）與 Telegram Bot，提供照片查詢、天氣摘要與推播功能。

## 技術棧

- **Go 1.25**
- **Gin** — HTTP 框架
- **Air** — Hot reload 工具
- **Docker Compose** — 開發環境
- **Telegram Bot API** — 訊息互動與推播
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
│   │   └── weather.go          # 天氣 API
│   ├── service/
│   │   ├── telegram.go         # Telegram bot、排程、推播邏輯
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
```

> 到 [unsplash.com/oauth/applications](https://unsplash.com/oauth/applications) 取得 Access Key。
>
> 到 [中央氣象署開放資料平台](https://opendata.cwa.gov.tw/) 取得授權碼（Authorization）。

備註：天氣金鑰同時支援 `CWB_API_KEY` 或 `CWA_API_KEY`。

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

| Method | Path               | 說明                                       |
| ------ | ------------------ | ------------------------------------------ |
| GET    | `/`                | 健康檢查                                   |
| GET    | `/ping`            | 測試 Telegram 連線與發送訊息               |
| GET    | `/photos/random`   | 取得一張隨機照片                           |
| GET    | `/weather/current` | 取得指定城市 24 小時重點天氣（預設臺南市） |
| POST   | `/notify/test`     | 立即測試一次早安推播（共用排程內容）       |

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

# 測試推播（使用 .env 的 TELEGRAM_CHAT_ID）
curl -X POST "http://localhost:8080/notify/test"

# 測試推播（指定 chatId）
curl -X POST "http://localhost:8080/notify/test?chatId=5291868928"
```

## 天氣回應格式

`GET /weather/current` 回傳已清洗的 24 小時重點資料，主要欄位：

- `locationName`：城市名稱
- `timeRange`：查詢區間（from/to/hours）
- `periods`：各時段天氣摘要
- `current`：目前時段重點摘要

## Telegram Bot 指令

- `/menu`：顯示功能選單
- `/weather`：查詢預設城市天氣
- `/weather 城市名`：查詢指定城市
- `/setcity 城市名`：設定個人預設城市
- `/mycity`：查看目前預設城市
- `/image`：取得一張隨機照片

備註：目前預設城市存在記憶體（in-memory），服務重啟後會回到預設值 `臺南市`。

## 排程推播

- 每天台灣時間 `06:00` 會推播「早安圖片 + 24 小時天氣摘要」。
- 排程與 `POST /notify/test` 共用同一段推播邏輯，測試內容和正式排程一致。

## 安裝新套件

因為本地沒有安裝 Go，透過容器執行：

```bash
docker compose run --rm app go get <套件名稱>
```
