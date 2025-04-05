package crawler

type News struct {
	Title   string
	Link    string
	Cover   string
	Desc    string
	Content string
}

type NewsCrawler interface {
	FetchNewsList() ([]News, error)
	FetchNewsDetail(url string) (string, error)
}
