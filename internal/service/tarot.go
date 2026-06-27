package service

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/Yo0GuitarIT/yo0-backend/internal/model"
)

// 大阿爾克那 22 張
var majorArcana = []string{
	"The Fool", "The Magician", "The High Priestess", "The Empress",
	"The Emperor", "The Hierophant", "The Lovers", "The Chariot",
	"Strength", "The Hermit", "Wheel of Fortune", "Justice",
	"The Hanged Man", "Death", "Temperance", "The Devil",
	"The Tower", "The Star", "The Moon", "The Sun",
	"Judgement", "The World",
}

// 大牌英文名 → 中文名
var majorArcanaZh = map[string]string{
	"The Fool":           "愚者",
	"The Magician":       "魔術師",
	"The High Priestess": "女祭司",
	"The Empress":        "皇后",
	"The Emperor":        "皇帝",
	"The Hierophant":     "教皇",
	"The Lovers":         "戀人",
	"The Chariot":        "戰車",
	"Strength":           "力量",
	"The Hermit":         "隱者",
	"Wheel of Fortune":   "命運之輪",
	"Justice":            "正義",
	"The Hanged Man":     "倒吊人",
	"Death":              "死神",
	"Temperance":         "節制",
	"The Devil":          "惡魔",
	"The Tower":          "高塔",
	"The Star":           "星星",
	"The Moon":           "月亮",
	"The Sun":            "太陽",
	"Judgement":          "審判",
	"The World":          "世界",
}

var suits = []string{"Wands", "Cups", "Swords", "Pentacles"}
var ranks = []string{
	"Ace", "Two", "Three", "Four", "Five", "Six", "Seven",
	"Eight", "Nine", "Ten", "Page", "Knight", "Queen", "King",
}

// 牌組英文 → 中文
var suitZh = map[string]string{
	"Wands":     "權杖",
	"Cups":      "聖杯",
	"Swords":    "寶劍",
	"Pentacles": "錢幣",
}

// 點數英文 → 中文
var rankZh = map[string]string{
	"Ace": "一", "Two": "二", "Three": "三", "Four": "四", "Five": "五",
	"Six": "六", "Seven": "七", "Eight": "八", "Nine": "九", "Ten": "十",
	"Page": "侍者", "Knight": "騎士", "Queen": "皇后", "King": "國王",
}

// 牌組英文 → Wikimedia 檔名用的縮寫（Pentacles 特別縮成 Pents）
var suitFile = map[string]string{
	"Wands":     "Wands",
	"Cups":      "Cups",
	"Swords":    "Swords",
	"Pentacles": "Pents",
}

// tarotImageURL 用 Wikimedia 的 Special:FilePath 組出免 hash 的圖片網址
func tarotImageURL(fileName string) string {
	return "https://commons.wikimedia.org/wiki/Special:FilePath/" + fileName
}

// buildDeck 組出完整 78 張牌
func buildDeck() []model.Tarot {
	deck := make([]model.Tarot, 0, 78)

	for i, name := range majorArcana {
		// "The Fool" → "Fool"、"Wheel of Fortune" → "Wheel_of_Fortune"
		stem := strings.ReplaceAll(strings.TrimPrefix(name, "The "), " ", "_")
		fileName := fmt.Sprintf("RWS_Tarot_%02d_%s.jpg", i, stem) // 例如 RWS_Tarot_00_Fool.jpg

		deck = append(deck, model.Tarot{
			Name:   name,
			NameZh: majorArcanaZh[name],
			Arcana: "Major",
			Image:  tarotImageURL(fileName),
		})
	}

	for _, suit := range suits {
		for ri, rank := range ranks {
			fileName := fmt.Sprintf("%s%02d.jpg", suitFile[suit], ri+1) // 例如 Wands01.jpg

			deck = append(deck, model.Tarot{
				Name:   rank + " of " + suit,
				NameZh: suitZh[suit] + rankZh[rank], // 例如「權杖一」
				Arcana: "Minor",
				Suit:   suit,
				Image:  tarotImageURL(fileName),
			})
		}
	}

	return deck // len(deck) == 78
}

// 整副牌只組一次，存起來重複用
var tarotDeck = buildDeck()

// GetRandomTarot 隨機抽一張牌（含隨機正逆位）
// 回傳：牌、HTTP 狀態碼、錯誤 — 對齊專案其它 service 的簽章
func GetRandomTarot() (*model.Tarot, int, error) {
	card := tarotDeck[rand.IntN(len(tarotDeck))] // 0 ~ 77
	card.Reversed = rand.IntN(2) == 1            // 50% 逆位

	return &card, 200, nil
}
