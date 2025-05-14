package main

import (
	"fmt"
	"github.com/Nolions/CricketNewsCrawler/crawler"
	"log"
)

func main() {
	cricbuzz := crawler.NewSportSkeeda()

	newsList, err := cricbuzz.FetchNewsList()
	if err != nil {
		log.Fatalf("爬取新聞失敗: %v", err)
	}

	//顯示結果
	for i, n := range newsList {
		fmt.Printf("新聞 %d: %s\nLink: %s\nDesc: %s\nCover: %s\n\n", i+1, n.Title, n.Link, n.Desc, n.Cover)
	}

	//news := crawler.News{}

	//err := cricbuzz.FetchNewsDetail(" https://www.cricbuzz.com/cricket-news/134344/ipl-2025-to-resume-on-may-17-final-on-june-3", &news)
	//
	//fmt.Println(err)
	//fmt.Println(news)
}
