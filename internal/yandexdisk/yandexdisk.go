package yandexdisk

import (
	"encoding/json"
	"net/http"
)

type DownloadResponse struct {
	Href string `json:"href"`
}

func GetDownloadLink(publicKey string) (string, error) {
	api := "https://cloud-api.yandex.net/v1/disk/public/resources/download?public_key=" + publicKey

	resp, err := http.Get(api)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Href, nil
}
