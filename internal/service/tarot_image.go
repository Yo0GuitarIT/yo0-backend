package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"time"
)

// fetchRotatedImage 下載牌面圖、旋轉 180 度（逆位用），回傳 JPEG bytes。
// 為了縮小流量與加快解碼，下載時用 Wikimedia 的 ?width 參數取較小的縮圖。
func fetchRotatedImage(imageURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", imageURL+"?width=600", nil)
	if err != nil {
		return nil, err
	}
	// Wikimedia 會擋掉沒有 User-Agent 的請求（預設的 Go-http-client 會被 403）
	req.Header.Set("User-Agent", "yo0-backend/1.0 (https://github.com/Yo0GuitarIT/yo0-backend)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下載圖片失敗，HTTP %d", resp.StatusCode)
	}

	src, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解碼圖片失敗: %w", err)
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 來源 (x,y) → 目的 (w-1-x, h-1-y)：上下 + 左右翻轉 = 旋轉 180°
			dst.Set(w-1-x, h-1-y, src.At(bounds.Min.X+x, bounds.Min.Y+y))
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("編碼圖片失敗: %w", err)
	}
	return buf.Bytes(), nil
}
