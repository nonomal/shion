package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/samber/lo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []Asset
}

type Asset struct {
	Size int
	Name string
	Url  string `json:"browser_download_url"`
}

func CheckUpgrade() (needUpgrade bool, tagName string, asset Asset, err error) {
	res, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", author, appName))
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	var data Release
	err = json.Unmarshal(body, &data)
	tagName = data.TagName
	if err != nil {
		return
	}
	asset, _ = lo.Find(data.Assets, func(asset Asset) bool {
		return asset.Name == "shion.exe"
	})
	newVersion := data.TagName[1:]
	needUpgrade = semver.New(version).LessThan(*semver.New(newVersion))
	return
}

func downloadProgress(done chan int64, path string, total int64) (err error) {
	stop := false
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				return err
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			println(size, total)

			runtime.EventsEmit(app.ctx, "upgrading", size, total)
		}

		if stop {
			break
		}
		time.Sleep(time.Second / 60)
	}
	return
}

func download(asset Asset) (path string, err error) {
	logger.Sugar().Info("download...", "url", asset.Url)
	dir := GetDownloadConfigDir()
	path = filepath.Join(dir, asset.Name)

	out, err := os.Create(path)
	if err != nil {
		return
	}
	defer out.Close()

	resp, err := http.Get(asset.Url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	done := make(chan int64)
	go downloadProgress(done, path, int64(asset.Size))

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return
	}
	done <- written

	logger.Info("download finish")
	return
}

func replace(assetExe string) (currentExe string, err error) {
	currentExe, err = os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(currentExe)
	oldExe := filepath.Join(dir, "shion-old.exe")
	err = os.Rename(currentExe, oldExe)
	if err != nil {
		return
	}
	err = moveFile(assetExe, currentExe)
	return
}

func moveFile(sourcePath, destPath string) (err error) {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return
	}
	err = os.Remove(sourcePath)
	if err != nil {
		return
	}
	return
}

func DeleteUpgradeTemp() {
	currentExe, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(currentExe)
	oldExe := filepath.Join(dir, "shion-old.exe")
	os.Remove(oldExe)
	download := GetDownloadConfigDir()
	os.Remove(download)
}

func Upgrade(asset Asset) {
	assetExe, err := download(asset)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	currentExe, err := replace(assetExe)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	exec.Command(currentExe, "-oldVersion", version).Start()
	os.Exit(0)
}
