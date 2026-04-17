package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type YandexResp struct {
	Href string `json:"href"`
}

func getDownloadLink(publicKey string) (string, error) {
	api := "https://cloud-api.yandex.net/v1/disk/public/resources/download?public_key=" + publicKey

	resp, err := http.Get(api)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data YandexResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Href, nil
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fullPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fullPath, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)

		outFile, err := os.Create(fullPath)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	publicKey := "https://disk.yandex.ru/d/eCzV6zSDSG-l_A"

	fmt.Println("Подготовка к скачиванию...")

	link, err := getDownloadLink(publicKey)
	if err != nil {
		fmt.Println("API error:", err)
		return
	}

	fmt.Println("Скачивание архива с модами...")

	if err := downloadFile(link, "mods.zip"); err != nil {
		fmt.Println("Ошибка скачивания:", err)
		return
	}

	fmt.Println("Распаковка в temp_mods...")

	os.RemoveAll("temp_mods")

	if err := unzip("mods.zip", "temp_mods"); err != nil {
		fmt.Println("Ошибка распаковки:", err)
		return
	}

	newModsPath := filepath.Join("temp_mods/moonawak3", "mods")

	// 🔥 ВАЖНАЯ ПРОВЕРКА
	if !pathExists(newModsPath) {
		fmt.Println("❌ Папка mods не найдена в скаченном архиве.")

		os.RemoveAll("temp_mods")
		os.Remove("mods.zip")
		return
	}

	fmt.Println("✔ Распаковка прошла успешно. Происходит замена модов...")

	// только теперь удаляем старые
	os.RemoveAll("mods")

	err = os.Rename(newModsPath, "mods")
	if err != nil {
		fmt.Println("Ошибка замены модов:", err)
		return
	}

	fmt.Println("Удаление временных файлов...")

	os.RemoveAll("temp_mods")
	os.Remove("mods.zip")

	fmt.Println("Выполнено!")

}
