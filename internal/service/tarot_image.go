package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // 註冊 PNG decoder，讓 image.Decode 能解析 JoJo wiki 的 .png 圖片
	"net/http"
	"time"
)

// fetchRotatedImage 下載牌面圖、旋轉 180 度（逆位用），回傳 JPEG bytes。
// Wikimedia 支援 ?width=600 縮圖參數；JoJo wiki 為靜態 CDN，不加參數。
func fetchRotatedImage(imageURL string) ([]byte, error) {
	requestURL := imageURL
	if len(imageURL) > 20 && imageURL[:20] == "https://commons.wiki" {
		requestURL = imageURL + "?width=600"
	}
	req, err := http.NewRequest("GET", requestURL, nil)
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
