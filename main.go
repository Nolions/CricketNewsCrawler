package main

import (
	"cricketNewsCrawler/crawler"
	"fmt"
	"log"
)

func main() {
	cricbuzz := crawler.NewCricbuzz()
	
	newsList, err := cricbuzz.FetchNewsList()
	if err != nil {
		log.Fatalf("爬取新聞失敗: %v", err)
	}

	// 顯示結果
	for i, n := range newsList {
		fmt.Printf("新聞 %d: %s\nLink: %s\nDesc: %s\nCover: %s\n\n", i+1, n.Title, n.Link, n.Desc, n.Cover)
	}

	content, _ := cricbuzz.FetchNewsDetail("https://www.cricbuzz.com/cricket-news/133952/shreyas-iyer-30-from-spin-hitter-to-modern-t20-marauder")

	fmt.Println("內文：", content)
}
