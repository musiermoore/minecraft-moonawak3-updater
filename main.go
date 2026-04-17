package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"moonawak3-minecraft/internal/archive"
	"moonawak3-minecraft/internal/console"
	"moonawak3-minecraft/internal/downloader"
	"moonawak3-minecraft/internal/fsutil"
	"moonawak3-minecraft/internal/mods"
	"moonawak3-minecraft/internal/selector"
	"moonawak3-minecraft/internal/yandexdisk"
)

const (
	publicKey       = "https://disk.yandex.ru/d/eCzV6zSDSG-l_A"
	archivePath     = "mods.zip"
	tempArchiveDir  = "temp_mods"
	tempSelectedDir = "temp_selected_mods"
	targetModsDir   = "mods"
)

func main() {
	defer console.PauseBeforeExit()

	fmt.Println("Подготовка к скачиванию...")

	link, err := yandexdisk.GetDownloadLink(publicKey)
	if err != nil {
		fmt.Println("API error:", err)
		return
	}

	fmt.Println("Скачивание архива с модами...")

	if err := downloader.DownloadFile(link, archivePath); err != nil {
		fmt.Println("Ошибка скачивания:", err)
		return
	}

	fmt.Println("Распаковка в temp_mods...")

	os.RemoveAll(tempArchiveDir)

	if err := archive.Unzip(archivePath, tempArchiveDir); err != nil {
		fmt.Println("Ошибка распаковки:", err)
		return
	}

	archiveRoot := filepath.Join(tempArchiveDir, "moonawak3")
	newModsPath := filepath.Join(archiveRoot, targetModsDir)

	if !fsutil.DirExists(newModsPath) {
		fmt.Println("❌ Папка mods не найдена в скаченном архиве.")

		os.RemoveAll(tempArchiveDir)
		os.Remove(archivePath)
		return
	}

	fmt.Println("✔ Распаковка прошла успешно.")

	selected, err := selector.SelectFilesForMods(archiveRoot)
	if err != nil {
		if errors.Is(err, selector.ErrExitRequested) {
			os.RemoveAll(tempArchiveDir)
			os.RemoveAll(tempSelectedDir)
			os.Remove(archivePath)
			console.DisablePauseBeforeExit()
			return
		}
		fmt.Println("Ошибка выбора файлов:", err)
		return
	}

	if len(selected) == 0 {
		fmt.Println("Ничего не выбрано. Замена модов отменена.")
		return
	}

	fmt.Println("Подготовка новой папки mods...")

	if err := mods.BuildFolder(archiveRoot, tempSelectedDir, selected); err != nil {
		fmt.Println("Ошибка подготовки модов:", err)
		os.RemoveAll(tempSelectedDir)
		return
	}

	fmt.Println("Происходит замена модов...")

	os.RemoveAll(targetModsDir)

	err = os.Rename(tempSelectedDir, targetModsDir)
	if err != nil {
		fmt.Println("Ошибка замены модов:", err)
		os.RemoveAll(tempSelectedDir)
		return
	}

	fmt.Println("Удаление временных файлов...")

	os.RemoveAll(tempArchiveDir)
	os.RemoveAll(tempSelectedDir)
	os.Remove(archivePath)

	fmt.Println("Выполнено!")
}
