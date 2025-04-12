package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	manhwaList  []string
	mangaUrl    string
	resp        *http.Response
	err         error
	manhwaSrc   string
	mangaPriSrc string
	mangaSecSrc string
)

type Manga struct {
	mangaName         string
	mangaType         string
	latestChapterRead int
}

func init() {
	godotenv.Load()
	manhwaSrc = os.Getenv("MANHWA_SRC_URL")
	manhwaList = []string{
		os.Getenv("DRAGON_SLAYER_MANGA"),
		os.Getenv("KNIGHT_KING_WHO_RETURNED_WITH_GOD_MANGA"),
	}
	mangaPriSrc = os.Getenv("MANGA_PRI_SRC_URL")
	mangaSecSrc = os.Getenv("MANGA_SEC_SRC_URL")
}

func main() {
	mangaList := []Manga{
		{
			mangaName:         "A Dragonslayer's Peerless Regression",
			mangaType:         "manhwa",
			latestChapterRead: 33,
		},
		{
			mangaName:         "The Knight King Who Returned with a God",
			mangaType:         "manhwa",
			latestChapterRead: 107,
		},
		{
			mangaName:         "Kingdom",
			mangaType:         "manga",
			latestChapterRead: 345,
		},
	}

	var wg sync.WaitGroup
	wg.Add(len(mangaList))
	for _, manga := range mangaList {
		go func(manga Manga) {
			defer wg.Done()
			newChapterPublished := false
			if manga.mangaType == "manhwa" {
				newChapterPublished = processManhwa(manga)
			} else if manga.mangaType == "manga" {
				newChapterPublished = processManga(manga)
			}

			if newChapterPublished {
				go notifyUser(manga)
			}
		}(manga)
	}

	wg.Wait()
}

func notifyUser(manga Manga) {
	fmt.Println("manga: ", manga.mangaName)
	fmt.Println("chapter: ", manga.latestChapterRead+1)
	// TODO
}

func processManhwa(manga Manga) bool {
	mangaAvailable, manhwaName := isManhwaPresent(manga.mangaName)
	if !mangaAvailable {
		return false
	}

	return newMangaChapterAvailable(manhwaSrc, manga.latestChapterRead, manhwaName)
}

func processManga(manga Manga) bool {
	manga.mangaName = formatMangaName(manga.mangaName)
	return newMangaChapterAvailable(mangaPriSrc, manga.latestChapterRead, manga.mangaName) || newMangaChapterAvailable(mangaSecSrc, manga.latestChapterRead, manga.mangaName)
}

func newMangaChapterAvailable(mangaSrc string, latestChapterRead int, mangaName string) bool {
	mangaUrl = formatMangaUrl(mangaSrc, latestChapterRead, mangaName)
	resp, err = http.Get(mangaUrl)

	if isValidResponse(resp, err) {
		// latest chapter found
		return true
	}
	return false
}

func formatManhwaName(mangaName string) string {
	return strings.ReplaceAll(strings.Join(strings.Split(strings.ToLower(mangaName), " "), "-"), "'", "")
}

func formatMangaName(mangaName string) string {
	return strings.Join(strings.Split(strings.ToLower(mangaName), " "), "-")
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
