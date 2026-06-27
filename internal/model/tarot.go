package model

type Tarot struct {
	Name     string `json:"name"`     // 牌名，例如 "The Fool"
	NameZh   string `json:"nameZh"`   // 中文名，例如 "愚者"
	Arcana   string `json:"arcana"`   // "Major" 或 "Minor"
	Suit     string `json:"suit"`     // 小牌的牌組（大牌為空）
	Image    string `json:"image"`    // RWS 牌面圖片網址（Wikimedia，公有領域）
	Meaning  string `json:"meaning"`  // 依正/逆位對應的牌義（一句話）
	Reversed bool   `json:"reversed"` // true = 逆位
}