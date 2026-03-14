package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetRandomPhoto() (map[string]interface{}, int, error) {
	accessKey := os.Getenv("UNSPLASH_ACCESS_KEY")

	req, err := http.NewRequest("GET", "https://api.unsplash.com/photos/random", nil)
	if err != nil {
		return nil, 500, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", accessKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 500, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 500, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, 500, err
	}

	return result, resp.StatusCode, nil
}
