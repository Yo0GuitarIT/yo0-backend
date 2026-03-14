# yo0-backend

Go 後端服務，整合 Unsplash API，提供隨機照片查詢功能。

## 技術棧

- **Go 1.25**
- **Gin** — HTTP 框架
- **Air** — Hot reload 工具
- **Docker Compose** — 開發環境

## 專案結構

```
yo0-backend/
├── cmd/
│   └── main.go                 # 程式入口
├── internal/
│   ├── handler/
│   │   └── photo.go            # HTTP 請求處理
│   ├── service/
│   │   └── unsplash.go         # Unsplash API 邏輯
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
```

> 到 [unsplash.com/oauth/applications](https://unsplash.com/oauth/applications) 取得 Access Key。

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

| Method | Path             | 說明             |
| ------ | ---------------- | ---------------- |
| GET    | `/`              | 健康檢查         |
| GET    | `/photos/random` | 取得一張隨機照片 |

### 範例

```bash
curl http://localhost:8080/photos/random
```

## 安裝新套件

因為本地沒有安裝 Go，透過容器執行：

```bash
docker compose run --rm app go get <套件名稱>
```
