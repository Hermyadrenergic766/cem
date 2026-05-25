package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// updateCheckCache — son kontrol zamanı ve görülen son sürüm.
type updateCheckCache struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
}

const updateCheckInterval = 1 * time.Hour

// checkUpdateNotice — her cem/cemi/cemir invokasyonunun başında çağrılır.
// Cache TTL doluysa arka planda GitHub API'sini sorgulayıp dosyayı günceller;
// mevcut cache yeni sürüm gösteriyorsa renkli bildirim basar.
func checkUpdateNotice() {
	cache := loadUpdateCheckCache()
	stale := time.Since(cache.LastCheck) > updateCheckInterval

	if stale {
		// Non-blocking refresh — bu invokasyonu yavaşlatmasın
		go refreshUpdateCheck()
	}

	if cache.LatestVersion != "" && version != "dev" && semverLess(version, cache.LatestVersion) {
		printUpdateNotice(cache.LatestVersion)
	}
}

func printUpdateNotice(latest string) {
	// Sarı/turuncu vurgu — dikkat çeksin, çıktıyı boğmasın
	fmt.Printf("  %s %s available (current: %s) · run %s\n",
		styleWarn.Render("🔔 new version"),
		styleBold.Render(latest),
		styleDim.Render(version),
		styleBold.Render("cem update"))
}

func updateCheckPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cem", "update-check.json")
}

func loadUpdateCheckCache() updateCheckCache {
	var c updateCheckCache
	data, err := os.ReadFile(updateCheckPath())
	if err != nil {
		return c
	}
	_ = json.Unmarshal(data, &c)
	return c
}

func saveUpdateCheckCache(c updateCheckCache) {
	data, err := json.Marshal(c)
	if err != nil {
		return
	}
	dir := filepath.Dir(updateCheckPath())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	_ = os.WriteFile(updateCheckPath(), data, 0o600)
}

func refreshUpdateCheck() {
	latest, err := fetchLatestVersion()
	if err != nil {
		return
	}
	saveUpdateCheckCache(updateCheckCache{
		LastCheck:     time.Now(),
		LatestVersion: latest,
	})
}
