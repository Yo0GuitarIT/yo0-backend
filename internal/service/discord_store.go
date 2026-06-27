package service

import (
	"sync"
	"time"
)

// lastDrawDate 以 Discord 使用者 ID 為 key，記錄該使用者最後一次抽牌的日期
//（台灣時區、格式 2006-01-02）。屬於記憶體儲存，服務重啟後會重置。
var (
	drawMu       sync.Mutex
	lastDrawDate = map[string]string{}
)

// todayInTaipei 回傳台灣時區的今日日期字串，作為「當天」的判斷依據。
func todayInTaipei() string {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.Now().Format("2006-01-02")
	}
	return time.Now().In(loc).Format("2006-01-02")
}

// tryMarkDraw 嘗試把使用者標記為「今天已抽」。
// 回傳 true 表示這是今天第一次抽（允許）；false 表示今天已經抽過。
// 檢查與標記在同一把鎖內完成，避免連續快速點擊造成的競態。
func tryMarkDraw(userID string) bool {
	drawMu.Lock()
	defer drawMu.Unlock()

	today := todayInTaipei()
	if lastDrawDate[userID] == today {
		return false
	}
	lastDrawDate[userID] = today
	return true
}
