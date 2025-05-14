# CricketNewsCrawler

爬取下面四個網站的板球新聞

1. [cricbuzz](https://www.cricbuzz.com/cricket-news)
2. [espncricinfo](https://www.espncricinfo.com/cricket-news)
3. [ndtv](https://sports.ndtv.com/cricket/news)
4. [sportskeeda](https://www.sportskeeda.com/cricket)
5. [bcci](https://www.bcci.tv/international/men/news)

## use

***取得/下載套件***

```bash
go get github.com/Nolions/CricketNewsCrawler@latest
```

### code illustration

### Method

***取得新聞列表***

FetchNewsList() ([]News, error)

***取得單一新聞內文***

FetchNewsDetail(url string, news *News) error

| 參數   | 說明       |
|------|----------|
| url  | 新聞詳細業面連結 |
| news |          |

> 不同物件的FetchNewsList()與FetchNewsDetail()不能混用，
> EX: SportSkeeda.FetchNewsList()取得的新聞詳情頁的連結就只能透過SportSkeeda.FetchNewsDetail()取得取新聞內文

#### News 結構

| 參數      | 型態        | 說明       |
|---------|-----------|----------|
| Title   | string    | 新聞標題     |
| Link    | string    | 新聞詳細頁面連結 |
| Cover   | string    | 新聞封面     |
| Desc    | string    | 新聞描述     |
| PubDate | time.Time | 新聞發布日期   |

 
### example

***建立Cricbuzz爬蟲實例***

```go
cricbuzz := crawler.NewCricbuzz()
```

***取得Cricbuzz網站中新聞列表***

```go
newsList, err := cricbuzz.FetchNewsList()
if err != nil {
    log.Fatalf("爬取新聞失敗: %v", err)
} else {
//顯示結果
    for i, n := range newsList {
        fmt.Printf("新聞 %d: %s\nLink: %s\nDesc: %s\nCover: %s\n\n", i+1, n.Title, n.Link, n.Desc, n.Cover)
    }
}


```

***取得指定頁面新聞內文***

```go
news := crawler.News{}
err := cricbuzz.FetchNewsDetail(" https://www.cricbuzz.com/cricket-news/134344/ipl-2025-to-resume-on-may-17-final-on-june-3", &news)
if err != nil {
    log.Fatalf("爬取新聞內文失敗: %v", err)
} else {
    fmt.Println(news)
}

```

## Status

- [x] [cricbuzz](https://www.cricbuzz.com/cricket-news)
- [x] [espncricinfo](https://www.espncricinfo.com/cricket-news)
- [x] [ndtv](https://sports.ndtv.com/cricket/news)
- [x] [sportskeeda](https://www.sportskeeda.com/cricket)
- [x] [bcci](https://www.bcci.tv/international/men/news)
