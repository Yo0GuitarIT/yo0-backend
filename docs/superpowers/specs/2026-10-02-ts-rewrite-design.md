# yo0-backend TypeScript 改寫設計

- 日期：2026-10-02
- 狀態：待審閱
- 分支：`ts-rewrite`

## 1. 目標與範圍

### 1.1 為什麼改寫

1. 與前端統一語言，降低自己維護的切換成本。
2. 取用 TS / Node 生態系（discord.js、grammY 等功能更完整的套件），方便日後擴充。
3. 練習 TS 後端開發與測試。

### 1.2 成功標準

- **使用者可見功能不變**：Telegram 指令、Discord 抽牌互動、每日 06:00（台北）推播、HTTP 各端點的功能皆保留，Bot 訊息文字一字不差。
- **不合理的既有行為改用業界做法**（見第 3 節），不為了相容保留舊行為。
- 未被第 3 節改動的邏輯，必須通過從 Go 版錄下的快照測試（見第 7 節）。
- 部署於 Zeabur，以 Dockerfile 建置；Go 程式碼在同一個 PR 內完全移除。

### 1.3 不在範圍內

- 持久化儲存（用戶預設城市、每日抽牌紀錄仍為記憶體，重啟後重置）。
- 新增任何 Bot 指令或 HTTP 端點。
- 新舊版並行運作（Telegram long polling 同時只能有一個 consumer，必須一次切換）。

### 1.4 HTTP API 呼叫者

目前沒有前端或其他服務呼叫 HTTP API，只有開發者手動以 curl / Postman 測試，因此可以自由調整 API 合約。

## 2. 技術選型

| 用途 | Go 版 | TS 版 |
|---|---|---|
| Runtime | Go 1.25 | Node.js 22 LTS |
| HTTP 框架 | Gin | Hono（`@hono/node-server`） |
| Telegram | go-telegram-bot-api | grammY |
| Discord | discordgo | discord.js |
| 排程 | robfig/cron | croner |
| 圖片處理 | image 標準庫（手寫旋轉迴圈） | sharp |
| 設定驗證 | 手寫 `os.Getenv` | zod |
| Log | `log` 標準庫 | pino |
| 測試 | 無 | Vitest |
| Lint / 格式 | — | typescript-eslint + Prettier |
| Dev hot reload | air | tsx watch |
| 套件管理 | go mod | npm |

TypeScript 採 `strict: true`，ESM（`"type": "module"`），以 `tsc` 編譯至 `dist/`。

## 3. 相對 Go 版的刻意行為變更

| # | Go 版現況 | TS 版行為 |
|---|---|---|
| 1 | Telegram 任何非指令文字都回「不支援的指令」 | 一般文字不回應；僅未註冊的 `/指令` 回「不支援的指令，請使用 /menu 查看可用功能」 |
| 2 | CWA 任何失敗一律回 500 | 上游非 2xx → 502 `UPSTREAM_ERROR`；逾時 → 504 `UPSTREAM_TIMEOUT`；金鑰未設定 → 500 `CONFIG_ERROR` |
| 3 | Unsplash 原樣轉發上游狀態碼，錯誤 body 被硬解析成 Photo 回傳 | 上游非 2xx → 502；不轉發上游 body |
| 4 | 無天氣時段時省略 `current` 欄位；大牌 `suit` 為 `""` | 回應結構固定，無值以 `null` 表示：`current: null`、大牌 `suit: null` |
| 5 | 錯誤回應直接帶 `err.Error()`，外洩內部細節 | 回應只帶通用訊息與錯誤碼；詳細原因寫入 log |
| 6 | `GET /ping` 會發送 Telegram 訊息（GET 有副作用） | 改為 `POST /ping` |
| 7 | `/ping`、`/notify/test` 無驗證，任何人可觸發推播 | 需 `Authorization: Bearer <ADMIN_TOKEN>` |
| 8 | 未設 Telegram token 時仍嘗試連線而失敗；bot 未啟動時 `/ping` 回 500 | 未設 token 明確略過並 log；bot 未啟動時相關端點回 503 |
| 9 | Unsplash 請求無逾時；Telegram 發送錯誤被忽略 | 所有外部請求皆有逾時；所有發送失敗都寫 log |
| 10 | 無 graceful shutdown | 收到 SIGTERM / SIGINT 時依序關閉（見 5.7） |
| 11 | 環境變數在使用時才發現錯誤 | 啟動時以 zod 驗證格式，格式錯誤立即結束並印出原因 |

以下**刻意不改**：RFC3339 `+08:00` 時間格式、記憶體儲存、Discord 啟動時貼公告、Bot 所有訊息文字、早安推播內容與順序、HTTP 預設城市「臺南市」。

## 4. 專案結構

```
yo0-backend/
├── src/
│   ├── index.ts                 # 入口：驗證設定 → 啟動 HTTP / Telegram / Discord → 註冊關機流程
│   ├── config.ts                # zod schema、讀取環境變數
│   ├── logger.ts                # pino 實例
│   ├── errors.ts                # AppError 與子類別
│   ├── types.ts                 # Tarot / Photo / WeatherSummary 等型別
│   ├── lib/
│   │   ├── fetch-json.ts        # fetch + 逾時 + 錯誤轉換
│   │   └── time.ts              # 台北時區工具：todayInTaipei、toRFC3339Taipei、formatForDisplay
│   ├── http/
│   │   ├── app.ts               # createApp(deps)：建立 Hono app 與路由，不 listen
│   │   ├── auth.ts              # Bearer token middleware
│   │   └── handlers/            # ping / photos / weather / notify / tarot
│   ├── services/
│   │   ├── weather.ts           # fetchWeather（打 CWA）+ extractWeatherSummary（純函式）
│   │   ├── unsplash.ts
│   │   └── tarot/
│   │       ├── deck.ts          # buildDeck（純函式）、drawTarot
│   │       ├── meanings.ts
│   │       ├── image.ts         # fetchRotatedImage（sharp）
│   │       └── data/            # tarot_cards.json、tarot_meanings.json
│   ├── telegram/
│   │   ├── bot.ts               # createTelegramBot、指令註冊、bot.catch
│   │   ├── commands.ts          # 各指令 handler
│   │   ├── formatter.ts         # normalizeCityName、formatWeatherMessage、formatTarotCaption（純函式）
│   │   ├── push.ts              # sendMorningPush
│   │   ├── scheduler.ts         # croner 排程
│   │   └── store.ts             # 用戶預設城市 Map
│   └── discord/
│       ├── bot.ts               # client 建立、啟動公告、互動事件分派
│       ├── tarot-interaction.ts # 抽牌互動流程、按鈕元件
│       └── store.ts             # 每日抽牌紀錄 Map
├── test/
│   ├── fixtures/                # 從 Go 版錄下的 golden 檔
│   └── **/*.test.ts
├── .github/workflows/ci.yml
├── Dockerfile
├── docker-compose.yml
├── package.json / tsconfig.json / eslint.config.js / .prettierrc
└── README.md
```

設計原則：
- **I/O 與純邏輯分離**：天氣清洗、訊息格式化、牌組組成皆為不打網路的純函式，可直接以 fixture 測試。
- **依賴注入取代全域單例**：HTTP handler 與推播透過參數取得 Telegram bot（可能為 `null`），測試時可替換為 fake。
- JSON 資料以 `import ... with { type: "json" }` 載入，取代 `go:embed`。

## 5. 元件設計

### 5.1 設定（`config.ts`）

| 變數 | 必填 | 說明 |
|---|---|---|
| `PORT` | 否 | 預設 `8080` |
| `LOG_LEVEL` | 否 | 預設 `info` |
| `ADMIN_TOKEN` | 否 | 未設定時，受保護端點一律回 503 `SERVICE_UNAVAILABLE`（預設拒絕） |
| `TELEGRAM_BOT_TOKEN` | 否 | 未設定則不啟動 Telegram（含排程） |
| `TELEGRAM_CHAT_ID` | 否 | 必須是整數；排程、`/ping`、`/notify/test` 的預設對象。未設定時不建立排程 |
| `CWB_API_KEY` / `CWA_API_KEY` | 否 | 擇一，前者優先；未設定時呼叫天氣丟 `ConfigError` |
| `UNSPLASH_ACCESS_KEY` | 否 | 未設定時呼叫照片丟 `ConfigError` |
| `DISCORD_BOT_TOKEN` | 否 | 未設定則不啟動 Discord |
| `DISCORD_CHANNEL_IDS` | 否 | 逗號分隔、去除空白與空值；未設定時退而讀 `DISCORD_CHANNEL_ID` |

規則：未設定的選填變數只會停用對應功能並寫一筆 `warn` log；**有設定但格式錯誤**（例如 `TELEGRAM_CHAT_ID=abc`）則啟動失敗並以非零結束碼結束。

### 5.2 錯誤模型（`errors.ts`）

| 類別 | HTTP 狀態 | `code` |
|---|---|---|
| `BadRequestError` | 400 | `BAD_REQUEST` |
| `UnauthorizedError` | 401 | `UNAUTHORIZED` |
| `ConfigError` | 500 | `CONFIG_ERROR` |
| `UpstreamError` | 502 | `UPSTREAM_ERROR` |
| `ServiceUnavailableError` | 503 | `SERVICE_UNAVAILABLE` |
| `UpstreamTimeoutError` | 504 | `UPSTREAM_TIMEOUT` |
| 其他未預期錯誤 | 500 | `INTERNAL_ERROR` |
| 路由不存在 | 404 | `NOT_FOUND` |

所有錯誤回應格式：

```json
{ "error": { "code": "UPSTREAM_ERROR", "message": "天氣服務暫時無法使用" } }
```

`app.onError` 負責轉換並寫 log（含原始錯誤與 stack）；`message` 為給呼叫者看的通用中文說明，不含上游細節。

`fetchJson(url, { timeoutMs, headers })`：使用原生 `fetch` 與 `AbortSignal.timeout()`。逾時 → `UpstreamTimeoutError`；網路錯誤或非 2xx → `UpstreamError`（log 記錄上游狀態碼）。逾時：CWA、Unsplash 10 秒；牌面圖片 15 秒。

### 5.3 HTTP API

| 方法與路徑 | 驗證 | 成功回應 | 可能錯誤 |
|---|---|---|---|
| `GET /` | — | 200 `{"message":"yo0-backend is running!"}` | — |
| `POST /ping` | Bearer | 200 `{"message":"pong","telegram":"sent"}`，並發送「👋 測試訊息！yo0-backend 連線正常。」到 `TELEGRAM_CHAT_ID` | 401、503（bot 未啟動 / 未設 chat id / 未設 ADMIN_TOKEN）、502 |
| `GET /photos/random` | — | 200 `Photo` | 500、502、504 |
| `GET /weather/current?locationName=` | — | 200 `WeatherSummary`；`locationName` 預設「臺南市」 | 500、502、504 |
| `POST /notify/test?chatId=` | Bearer | 200 `{"message":"測試推播已送出","chatId":<number>}`；未帶 `chatId` 時用 `TELEGRAM_CHAT_ID` | 400（`chatId` 不是安全整數）、401、503、502、504 |
| `GET /tarot/random` | — | 200 `Tarot` | — |

Bearer 比對使用 `crypto.timingSafeEqual`。

回應型別：

```ts
type Photo = {
  id: string;
  description: string | null;
  alt_description: string | null;
  urls: { raw: string; full: string; regular: string; small: string; thumb: string };
};

type Tarot = {
  name: string;          // "The Fool"
  nameZh: string;        // "愚者"
  arcana: "Major" | "Minor";
  suit: "Wands" | "Cups" | "Swords" | "Pentacles" | null;
  image: string;
  meaning: string;       // 查無牌義時為 ""
  reversed: boolean;
};

type WeatherPeriod = {
  startTime: string;     // RFC3339，+08:00
  endTime: string;
  weather: string;
  rainProbability: string;
  minTempC: string;
  maxTempC: string;
  comfort: string;
};

type WeatherSummary = {
  locationName: string;
  timeRange: { from: string; to: string; hours: 24 };
  periods: WeatherPeriod[];
  current: WeatherPeriod | null;
};
```

### 5.4 Services

**天氣**（`services/weather.ts`）
- `fetchWeather(locationName, now = new Date())`：以台北時區計算 `timeFrom = now`、`timeTo = now + 24h`（格式 `YYYY-MM-DDTHH:mm:ss`），呼叫 `https://opendata.cwa.gov.tw/api/v1/rest/datastore/F-C0032-001`，帶 `Authorization`、`timeFrom`、`timeTo`、`locationName`（非空才帶）。
- `extractWeatherSummary(raw, locationName, from, to)` 為純函式，演算法與 Go 版一致：
  - 缺 `records` → 丟 `UpstreamError`（上游格式錯誤）。
  - `records.location` 為空 → 回傳 `periods: []`、`current: null`、`locationName` 用輸入值。
  - 取第一個 location；`locationName` 為空時退回輸入值。
  - 依 `startTime|endTime` 合併 `Wx / PoP / MinT / MaxT / CI` 到同一時段；時間以台北時區解析 `YYYY-MM-DD HH:mm:ss`，無法解析的項目略過。
  - 只保留 `end > from && start < to` 的時段，依 `start` 升冪排序。
  - `current` = 第一個時段，沒有則 `null`。
  - 欄位缺漏時以 `""` 填入（與 Go 零值一致）。

**Unsplash**（`services/unsplash.ts`）：`GET https://api.unsplash.com/photos/random`，header `Authorization: Client-ID <key>`，只挑出 `Photo` 型別定義的欄位回傳。

**塔羅牌**（`services/tarot/`）
- `buildDeck()`：22 張大牌（圖片來自 `tarot_cards.json` 的 JoJo 圖）+ 56 張小牌（Wikimedia `Special:FilePath/<Suit縮寫><兩位數>.jpg`，Pentacles 縮寫為 `Pents`），名稱與中文名規則同 Go 版。
- `drawTarot(random = Math.random)`：隨機抽一張，50% 逆位，帶入對應牌義。`random` 可注入以便測試。
- `fetchRotatedImage(url)`：以 `new URL(url).hostname` 判斷來源 —— `commons.wikimedia.org` 加 `?width=600`；`static.jojowiki.com` 加 `Referer: https://jojowiki.com/`；一律帶 `User-Agent: yo0-backend/1.0 (https://github.com/Yo0GuitarIT/yo0-backend)`。以 `sharp(buf).rotate(180).jpeg({ quality: 90 })` 輸出。

### 5.5 Telegram（grammY）

- 只有設了 `TELEGRAM_BOT_TOKEN` 才建立 bot，以 long polling 執行（`bot.start()`）。
- 指令：

| 指令 | 行為 |
|---|---|
| `/menu` | 回覆與 Go 版相同的指令清單文字 |
| `/weather [城市]` | 沒帶城市時用預設城市；經 `normalizeCityName` 後查詢；回 `formatWeatherMessage`；失敗回「❌ 取得天氣失敗，請稍後再試」 |
| `/setcity 城市` | 沒帶城市時回「請輸入城市名稱，例如：/setcity 臺南市」；成功回「✅ 已設定預設城市為：<城市>」 |
| `/mycity` | 「你的預設城市是：<城市>」 |
| `/image` | 傳送 Unsplash `urls.regular` 照片；失敗回「❌ 取得圖片失敗，請稍後再試」 |
| `/tarot` | 傳送牌面照片與 caption；失敗回「❌ 抽牌失敗，請稍後再試」 |
| 其他 `/指令` | 「不支援的指令，請使用 /menu 查看可用功能」 |
| 一般文字 | 不回應 |

- 塔羅照片：正位直接傳圖片 URL；逆位傳 `InputFile(旋轉後 bytes, "tarot.jpg")`，旋轉失敗則改傳原 URL。Caption：`🔮 {nameZh}（{name}）· {正位|逆位}\n{meaning}`。
- `formatWeatherMessage`、`normalizeCityName`（去除前後空白，「台」→「臺」）與 Go 版逐字一致；顯示時間格式 `YYYY-MM-DD HH:mm`（台北時區）。
- `sendMorningPush(bot, chatId)`：依序發送「☀️ 早安！今日天氣預報\n\n」+ 天氣訊息 → 隨機照片 → 塔羅照片；任一步失敗即中止並丟出錯誤。
- 排程：`new Cron("0 6 * * *", { timezone: "Asia/Taipei" })`，只在 bot 啟動且設了 `TELEGRAM_CHAT_ID` 時建立；成功與失敗都寫 log。
- `bot.catch` 統一記錄 handler 內未處理的錯誤；所有 `send*` 失敗都寫 log。
- 預設城市：`Map<number, string>`，預設值「臺南市」。

### 5.6 Discord（discord.js）

- 只有設了 `DISCORD_BOT_TOKEN` 才啟動；intents 只開 `Guilds`。
- `ready` 後在每個 `DISCORD_CHANNEL_IDS` 頻道貼出「恩雅婆婆開放占卜摟 🔮」與按鈕（`customId: draw_tarot`、label「🔮 抽牌」、Primary 樣式），訊息帶 `SuppressNotifications`；個別頻道失敗只寫 log。
- 按鈕互動流程（只處理 `customId === "draw_tarot"`）：
  1. `interaction.deferUpdate()`（失敗則 log 後結束）。
  2. 把被點的訊息按鈕改成「🔮 已抽過」、Secondary 樣式、disabled。
  3. `tryMarkDraw(userId)` 為同步函式（檢查與寫入之間沒有 `await`，單執行緒下不會有競態）；今天已抽過 → ephemeral follow-up「🔮 你今天已經占卜過囉，明天再來吧～」→ 補貼新按鈕 → 結束。
  4. 抽牌；失敗則在頻道送出「❌ 抽牌失敗，請稍後再試」（靜音）。
  5. 組 embed：description `**{顯示名稱}** 已為您占卜了`、title `🔮 {nameZh}（{name}）· {正位|逆位}`、有牌義時加入 field「牌義」。逆位附上 `tarot.jpg`，embed 圖片指向 `attachment://tarot.jpg`；旋轉失敗或正位時用原圖 URL。訊息靜音發送。
  6. 補貼新按鈕訊息「🔮 今日塔羅占卜，點下面的按鈕抽一張牌：」（靜音）。
- 顯示名稱：伺服器暱稱 → 帳號名稱 → 「神祕旅人」。
- 每日紀錄：`Map<userId, "YYYY-MM-DD">`（台北時區日期）。

### 5.7 啟動與關機（`index.ts`）

1. 解析設定（失敗即 `process.exit(1)`）。
2. 啟動 HTTP server（`PORT`）。
3. 依設定分別啟動 Telegram（含排程）與 Discord；任一個失敗只寫 log，不影響其他兩者。
4. 收到 `SIGTERM` / `SIGINT` 時依序：停止 cron → `bot.stop()` → `client.destroy()` → 關閉 HTTP server → 結束。設 10 秒強制結束的保險。

## 6. Docker 與開發環境

**正式 `Dockerfile`**（multi-stage）
- build stage：`node:22-bookworm-slim`，`npm ci` → `npm run build`（tsc）。
- runtime stage：`node:22-bookworm-slim`，只裝 production 依賴（`npm ci --omit=dev`），複製 `dist/`，`USER node`，`ENV NODE_ENV=production`，`EXPOSE 8080`，`CMD ["node", "dist/index.js"]`。
- sharp 在 linux glibc（x64 / arm64）有預先編譯好的版本，不需要額外的建置工具。

**開發用 `docker-compose.yml`**
- `node:22-bookworm-slim` image，掛載原始碼，`node_modules` 使用 named volume，`command: npm run dev`（`tsx watch src/index.ts`），`env_file: .env`，port `8080:8080`。
- 移除 `Dockerfile.dev`、`.air.toml`。

**npm scripts**：`dev`、`build`、`start`、`test`、`lint`、`typecheck`、`check`（= typecheck + lint + test）。

**CI**（`.github/workflows/ci.yml`）：push / PR 時執行 `npm ci && npm run check`。

## 7. 測試策略

### 7.1 階段 0：從 Go 版錄 golden fixtures（在刪除 Go 前完成）

在 `internal/service/` 新增 `golden_test.go`（package `service`，才能呼叫未匯出函式），執行 `go test ./internal/service -run Golden -update` 時將結果寫入 `test/fixtures/`：

| fixture | 內容 |
|---|---|
| `weather/cwa-*.json` + `weather/expected-*.json` | 一份真實 CWA 回應，以及手工建立的邊界情境：location 為空、欄位缺漏、時段超出範圍、時段順序打亂、時間格式錯誤。固定 `from` / `to`，錄下 `extractWeatherSummary` 的輸出 |
| `formatter/weather-message.json` | 多組 `WeatherSummary` → `formatWeatherMessage` 輸出（含 `current` 為 nil、欄位空白、`locationName` 為空） |
| `formatter/normalize-city.json` | `normalizeCityName` 的輸入與輸出 |
| `tarot/deck.json` | `buildDeck()` 完整 78 張 |
| `tarot/captions.json` | 指定牌與正逆位 → Telegram caption、Discord embed title 的文字 |

這個 Go 測試檔與 Go 程式碼一起刪除；fixtures 保留。

### 7.2 階段 1：TS 測試（Vitest）

- **純函式 vs fixtures**：天氣清洗、格式化、城市正規化、牌組、caption。第 3 節第 4 項的差異（大牌 `suit: ""` → `null`、沒有 `current` → `current: null`）在測試中以明確的轉換函式處理，不做寬鬆比對。
- **HTTP**：`createApp(deps).request()` 搭配 stub 的 `fetch` 與 fake Telegram bot，涵蓋 200 / 400 / 401 / 404 / 500 / 502 / 503 / 504 與錯誤回應格式、Bearer 驗證、`ADMIN_TOKEN` 未設定的情境。
- **設定**：選填變數未設定、格式錯誤、`CWB`/`CWA` 優先順序、`DISCORD_CHANNEL_IDS` 與舊名相容。
- **Telegram**：以 fake context 測試各指令回覆文字、一般文字不回應、未知指令提示、`sendMorningPush` 順序與中止行為。
- **Discord**：以 fake interaction 測試 `tryMarkDraw` 跨日重置、已抽過流程、正逆位 embed 組成、旋轉失敗時改用 URL。
- **圖片**：以小型測試圖驗證旋轉後像素位置正確，以及依來源加上不同的請求 header。

### 7.3 上線前手動驗收清單

- [ ] `GET /`、`GET /photos/random`、`GET /weather/current`（含 `locationName=高雄市`）、`GET /tarot/random`
- [ ] `POST /ping`、`POST /notify/test`：沒帶 token → 401；帶 token → 收到 Telegram 訊息
- [ ] Telegram：`/menu`、`/weather`、`/weather 台北市`、`/setcity`、`/mycity`、`/image`、`/tarot`（正逆位各一次）、未知指令、一般文字
- [ ] Discord：啟動公告、抽牌（正逆位）、同一天第二次抽牌、舊按鈕變灰、新按鈕補貼
- [ ] 隔天 06:00 收到早安推播

## 8. 部署切換（Zeabur）

1. 在 Zeabur 新增 `ADMIN_TOKEN`。
2. 合併 PR；Zeabur 偵測到根目錄的 `Dockerfile` 後自動建置部署。
3. 新舊容器交接期間可能短暫出現 Telegram 409；grammY 會自動重試，加上 graceful shutdown 可縮短衝突時間。
4. 執行 7.3 手動驗收清單。
5. 回滾：在 Zeabur 重新部署上一個版本，或 revert merge commit。

## 9. 實作順序

1. 階段 0：Go golden fixtures。
2. TS 專案骨架（package.json、tsconfig、lint、Vitest、CI）。
3. 基礎：config、logger、errors、fetch-json、time。
4. 純邏輯：天氣清洗、牌組、格式化（對 fixtures）。
5. Services I/O：weather、unsplash、tarot image。
6. HTTP app 與 handlers。
7. Telegram（指令、推播、排程）。
8. Discord。
9. index.ts 啟動與關機。
10. Dockerfile、docker-compose、README。
11. 刪除 Go 程式碼與相關檔案（`cmd/`、`internal/`、`go.mod`、`go.sum`、`.air.toml`、`Dockerfile.dev`）。
12. 部署切換與手動驗收。
