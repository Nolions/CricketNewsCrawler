package crawler

import "time"

type News struct {
	Title   string
	Link    string
	Cover   string
	Desc    string
	Content string
	PubDate time.Time
}

type NewsCrawler interface {
	FetchNewsList() ([]News, error)
	FetchNewsDetail(url string, news *News) error
}
