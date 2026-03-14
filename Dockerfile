# 第一階段：編譯
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# 先複製依賴檔案，利用 Docker 快取層避免每次重新下載套件
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼並編譯
COPY . .
RUN CGO_ENABLED=0 go build -o main ./cmd/main.go

# 第二階段：最小執行映像
FROM debian:bookworm-slim

WORKDIR /app

# 複製編譯好的執行檔
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["./main"]
