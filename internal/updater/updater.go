package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const appName = "moonawak3-minecraft"

var errNoRelease = errors.New("no GitHub release")

type release struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func HandleApplyUpdate(args []string) bool {
	if len(args) < 5 || args[1] != "--apply-update" {
		return false
	}

	newPath := args[2]
	targetPath := args[3]
	parentPID, _ := strconv.Atoi(args[4])

	if err := applyUpdate(newPath, targetPath, parentPID); err != nil {
		fmt.Println("Ошибка установки обновления:", err)
		os.Exit(1)
	}

	os.Exit(0)
	return true
}

func CheckAndInstall(currentVersion, owner, repo string) (bool, error) {
	latest, err := fetchLatestRelease(owner, repo)
	if err != nil {
		if errors.Is(err, errNoRelease) {
			return false, nil
		}
		return false, err
	}

	if !versionGreater(latest.TagName, currentVersion) {
		return false, nil
	}

	assetName := platformAssetName()
	asset, ok := findAsset(latest, assetName)
	if !ok {
		return false, fmt.Errorf("в релизе %s не найден файл %s", latest.TagName, assetName)
	}

	fmt.Printf("Доступна новая версия %s. Установить обновление? [y/N]: ", latest.TagName)
	answer := ""
	_, _ = fmt.Scanln(&answer)
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		return false, nil
	}

	fmt.Println("Скачивание обновления...")

	newPath, err := downloadAsset(asset)
	if err != nil {
		return false, err
	}

	if checksumAsset, ok := findAsset(latest, "checksums.txt"); ok {
		if err := verifyChecksum(checksumAsset.BrowserDownloadURL, asset.Name, newPath); err != nil {
			os.Remove(newPath)
			return false, err
		}
	}

	if err := installWithHelper(newPath); err != nil {
		os.Remove(newPath)
		return false, err
	}

	fmt.Println("Обновление скачано. Приложение перезапустится автоматически.")
	return true, nil
}

func fetchLatestRelease(owner, repo string) (release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return release{}, errNoRelease
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return release{}, fmt.Errorf("GitHub API вернул статус %s", resp.Status)
	}

	var latest release
	if err := json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return release{}, err
	}

	return latest, nil
}

func platformAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return appName + "-windows-" + runtime.GOARCH + ".exe"
	case "darwin":
		return appName + "-macos-" + runtime.GOARCH
	default:
		return appName + "-" + runtime.GOOS + "-" + runtime.GOARCH
	}
}

func findAsset(latest release, name string) (releaseAsset, bool) {
	for _, asset := range latest.Assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return releaseAsset{}, false
}

func downloadAsset(asset releaseAsset) (string, error) {
	resp, err := http.Get(asset.BrowserDownloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("скачивание %s вернуло статус %s", asset.Name, resp.Status)
	}

	ext := filepath.Ext(asset.Name)
	file, err := os.CreateTemp("", appName+"-*"+ext)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(file.Name())
		return "", err
	}

	return file.Name(), nil
}

func verifyChecksum(checksumURL, assetName, path string) error {
	resp, err := http.Get(checksumURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("скачивание checksums.txt вернуло статус %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	expected := checksumForAsset(string(body), assetName)
	if expected == "" {
		return fmt.Errorf("checksums.txt не содержит hash для %s", assetName)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum не совпадает для %s", assetName)
	}

	return nil
}

func checksumForAsset(content, assetName string) string {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := strings.TrimPrefix(fields[len(fields)-1], "*")
		if name == assetName || filepath.Base(name) == assetName {
			return fields[0]
		}
	}
	return ""
}

func installWithHelper(newPath string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	helperFile, err := os.CreateTemp("", appName+"-update-helper-*"+filepath.Ext(exePath))
	if err != nil {
		return err
	}
	helperPath := helperFile.Name()
	helperFile.Close()

	if err := copyFile(exePath, helperPath, 0755); err != nil {
		return err
	}

	cmd := exec.Command(helperPath, "--apply-update", newPath, exePath, strconv.Itoa(os.Getpid()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Start()
}

func applyUpdate(newPath, targetPath string, parentPID int) error {
	waitForParentExit(parentPID)

	if err := os.Chmod(newPath, 0755); err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < 60; attempt++ {
		if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := os.Rename(newPath, targetPath); err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return exec.Command(targetPath).Start()
	}

	return lastErr
}

func waitForParentExit(parentPID int) {
	time.Sleep(2 * time.Second)
}

func copyFile(src, dest string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func versionGreater(remote, current string) bool {
	remoteParts := versionParts(remote)
	currentParts := versionParts(current)

	for i := 0; i < 3; i++ {
		if remoteParts[i] > currentParts[i] {
			return true
		}
		if remoteParts[i] < currentParts[i] {
			return false
		}
	}

	return false
}

func versionParts(version string) [3]int {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	version = strings.Split(version, "-")[0]

	parts := strings.Split(version, ".")
	result := [3]int{}

	for i := 0; i < len(parts) && i < 3; i++ {
		n, _ := strconv.Atoi(parts[i])
		result[i] = n
	}

	return result
}
