package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
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
	ctx := context.Background()
	client, err := initializeRedis(ctx)
	if err != nil {
		panic(err)
	}
	vals, err := client.LRange(ctx, os.Getenv("MANHWA_LIST_KEY"), 0, -1).Result()
	if err != nil {
		panic(err)
	}
	manhwaList = vals
	manhwaSrc = os.Getenv("MANHWA_SRC_URL")
	mangaPriSrc = os.Getenv("MANGA_PRI_SRC_URL")
	mangaSecSrc = os.Getenv("MANGA_SEC_SRC_URL")
}

func initializeRedis(ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	// ping redis server
	pong, err := client.Ping(ctx).Result()
	fmt.Println(pong, err)
	if err != nil {
		return nil, err
	}
	return client, nil

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
			var url string
			if manga.mangaType == "manhwa" {
				newChapterPublished, url = processManhwa(manga)
			} else if manga.mangaType == "manga" {
				newChapterPublished, url = processManga(manga)
			}

			if newChapterPublished {
				go notifyUser(manga, url)
			}
		}(manga)
	}

	wg.Wait()
	select {}
}

func notifyUser(manga Manga, url string) {
	fmt.Println("manga: ", manga.mangaName)
	fmt.Println("chapter: ", manga.latestChapterRead+1)
	fmt.Println("url: ", url)
	// TODO
}

func processManhwa(manga Manga) (bool, string) {
	mangaAvailable, manhwaName := isManhwaPresent(manga.mangaName)
	if !mangaAvailable {
		return false, ""
	}

	return newMangaChapterAvailable(manhwaSrc, manga.latestChapterRead, manhwaName)
}

func processManga(manga Manga) (bool, string) {
	manga.mangaName = formatMangaName(manga.mangaName)
	priSrcAvailable, url := newMangaChapterAvailable(mangaPriSrc, manga.latestChapterRead, manga.mangaName)
	if priSrcAvailable {
		return true, url
	}

	return newMangaChapterAvailable(mangaSecSrc, manga.latestChapterRead, manga.mangaName)
}

func newMangaChapterAvailable(mangaSrc string, latestChapterRead int, mangaName string) (bool, string) {
	mangaUrl = formatMangaUrl(mangaSrc, latestChapterRead, mangaName)
	resp, err = http.Get(mangaUrl)

	if isValidResponse(resp, err) {
		// latest chapter found
		return true, resp.Request.URL.String()
	}
	return false, ""
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
