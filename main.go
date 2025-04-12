package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/robfig/cron"
)

var (
	manhwaList []string
	mangaUrl   string
	resp       *http.Response
	err        error
)

var manhwaSrc string

type Manga struct {
	mangaName         string
	mangaType         string
	latestChapterRead int
}

func init() {
	godotenv.Load()
	manhwaSrc = os.Getenv("ASURA_SCANS_URL")
	manhwaList = []string{
		os.Getenv("DRAGON_SLAYER_MANGA"),
	}
}

func main() {
	fmt.Println("hello world")
	cron.New()
	processManga(Manga{
		mangaName:         "A Dragonslayer's Peerless Regression",
		mangaType:         "manhwa",
		latestChapterRead: 33,
	})
	select {}
}

func processManga(manga Manga) {

	if manga.mangaType == "manhwa" {
		mangaAvailable, manhwaName := isManhwaPresent(manga.mangaName)
		if !mangaAvailable {
			return
		}
		mangaUrl = formatMangaUrl(manhwaSrc, manga.latestChapterRead, manhwaName)
		resp, err = http.Get(mangaUrl)
	} else if manga.mangaType == "manga" {

	}

	if isValidResponse(resp, err) {
		// latest chapter found
	}
}

func formatManhwaName(mangaName string) string {
	return strings.ReplaceAll(strings.Join(strings.Split(strings.ToLower(mangaName), " "), "-"), "'", "")
}

func isManhwaPresent(mangaName string) (bool, string) {
	mangaName = formatManhwaName(mangaName)
	for _, manhwa := range manhwaList {
		if strings.Contains(manhwa, mangaName) {
			return true, manhwa
		}
	}
	return false, ""
}

func formatMangaUrl(mangaUrl string, latestChapterRead int, mangaName string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			mangaUrl,
			"{chapterNumber}",
			strconv.Itoa(latestChapterRead+1),
		),
		"{mangaName}",
		mangaName,
	)
}

func isValidResponse(r *http.Response, e error) bool {
	if e != nil {
		fmt.Println(e.Error())
		return false
	}
	return r.StatusCode != http.StatusNotFound
}
