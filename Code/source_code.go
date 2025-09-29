package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	manifestFile  = "Vega.X_manifest.json"
	targetURL     = "https://github.com/1f0yt/community/releases/download/Vegax/Vega.X.apk"
	checkInterval = 60 * time.Second
)

var (
	spinnerFrames = []string{"|", "/", "-", "\\"}
	spinnerIndex  = 0
	spinnerActive = false
	spinnerLock   sync.Mutex
)

func fetchRemoteHash(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching remote file: %v", err)
	}
	defer resp.Body.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading remote file: %v", err)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func loadManifest() map[string]string {
	data := make(map[string]string)
	file, err := os.Open(manifestFile)
	if err != nil {
		return data
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&data)
	return data
}

func saveManifest(newHash string) {
	data := map[string]string{
		"last_checked": time.Now().Format("2006-01-02 15:04:05"),
		"release_hash": newHash,
	}

	file, err := os.Create(manifestFile)
	if err != nil {
		fmt.Printf("[!] Error saving manifest: %v\n", err)
		return
	}
	defer file.Close()

	json.NewEncoder(file).Encode(data)
	fmt.Println("[✓] Manifest updated.")
}

func spinner() {
	for {
		spinnerLock.Lock()
		active := spinnerActive
		spinnerLock.Unlock()

		if !active {
			break
		}

		fmt.Printf("\r[~] Checking %s", spinnerFrames[spinnerIndex%len(spinnerFrames)])
		spinnerIndex++
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Print("\r" + "                                        " + "\r")
}

func main() {
	fmt.Println("=== VegaX Version Scanner (By Megs Keller) ===")

	for {
		spinnerLock.Lock()
		spinnerActive = true
		spinnerLock.Unlock()

		go spinner()

		hash, err := fetchRemoteHash(targetURL)

		spinnerLock.Lock()
		spinnerActive = false
		spinnerLock.Unlock()

		time.Sleep(200 * time.Millisecond)

		if err != nil {
			fmt.Printf("[!] Failed to fetch remote hash: %v\n", err)
		} else {
			manifest := loadManifest()
			if manifest["release_hash"] != hash {
				fmt.Println("[+] New release detected!")
				fmt.Println("[!] Download the new version from, https://github.com/1f0yt/community/releases/download/Vegax/Vega.X.apk")
				saveManifest(hash)
			} else {
				fmt.Println("[-] No change detected.")
				fmt.Printf("[?] Last update was: %s\n", manifest["last_checked"])
			}
		}

		time.Sleep(checkInterval)
	}
}
