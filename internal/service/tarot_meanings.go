package service

import (
	_ "embed"
	"encoding/json"
	"log"
)

// tarotMeaningsJSON 在編譯時把牌義資料嵌入執行檔，執行時不需額外讀檔。
//
//go:embed tarot_meanings.json
var tarotMeaningsJSON []byte

// tarotMeaning 是單張牌的正/逆位牌義（各一句話）。
type tarotMeaning struct {
	Upright  string `json:"upright"`
	Reversed string `json:"reversed"`
}

// tarotMeanings 以英文牌名為 key，啟動時載入一次。
var tarotMeanings = loadTarotMeanings()

func loadTarotMeanings() map[string]tarotMeaning {
	var m map[string]tarotMeaning
	if err := json.Unmarshal(tarotMeaningsJSON, &m); err != nil {
		// 嵌入的資料若解析失敗代表 JSON 壞了，屬於開發期錯誤，直接讓它顯眼
		log.Panicf("載入 tarot_meanings.json 失敗: %v", err)
	}
	return m
}

// meaningFor 依正/逆位回傳對應牌義，查無資料時回空字串。
func meaningFor(name string, reversed bool) string {
	m, ok := tarotMeanings[name]
	if !ok {
		return ""
	}
	if reversed {
		return m.Reversed
	}
	return m.Upright
}
