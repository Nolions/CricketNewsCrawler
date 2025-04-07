package crawler

import (
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type BcciTV struct {
	Domain  string
	Headers map[string]string
}

func NewBcciTV() *BcciTV {
	headers := map[string]string{
		"sec-ch-ua":                 `"Chromium";v="133", "Not(A:Brand";v="99"`,
		"sec-ch-ua-mobile":          `?0`,
		"sec-ch-ua-platform":        `"macOS"`,
		"upgrade-insecure-requests": `1`,
		"user-agent":                `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36`,
	}

	return &BcciTV{
		Domain:  "https://www.bcci.tv",
		Headers: headers,
	}
}

// FetchNewsList
// 新聞列表
func (s *BcciTV) FetchNewsList() ([]News, error) {
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		return nil, err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		return nil, err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		return nil, fmt.Errorf("建立分頁失敗: %w", err)
	}

	url := "https://www.bcci.tv/international/men/news"
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status())
	if resp.Status() != http.StatusOK {
		return nil, fmt.Errorf("http Status is : %d", resp.Status())
	}

	els := page.Locator(".slick-track").First()
	newsEls := els.Locator("div.video-content.position-relative.slick-slide.slick-cloned")
	count, err := newsEls.Count()
	if err != nil {
		return nil, fmt.Errorf("獲取 <div class='slick-card'> 數量失敗: %w", err)
	}

	var newsList []News
	for i := 0; i < count; i++ {
		news, err := s.extractNewsItem(newsEls.Nth(i))
		if err != nil {
			log.Println("⚠️ 抓取單筆新聞失敗:", err)
			continue
		}
		newsList = append(newsList, news)
	}

	return newsList, nil
}

// 取得單筆新聞資訊
func (s *BcciTV) extractNewsItem(locator playwright.Locator) (News, error) {
	el := locator.Locator("a").First()
	title, err := el.GetAttribute("data-title")
	if err != nil {
		return News{}, fmt.Errorf("取得新聞標題錯誤: %w", err)
	}

	link, err := el.GetAttribute("href")
	if err != nil {
		return News{}, fmt.Errorf("取得新聞詳情連結錯誤: %w", err)
	}

	src, err := el.Locator("img").GetAttribute("src")
	if err != nil {
		return News{}, fmt.Errorf("取得新聞封面錯誤: %w", err)
	}

	timeEl := locator.Locator("ul li").First()
	timeStr, err := timeEl.TextContent()
	if err != nil {
		log.Println("取得新聞發布時間錯誤: ", err)
		timeStr = ""
	}

	t, _ := s.parseTime(timeStr)

	return News{
		Title:   title,
		Link:    link,
		Cover:   src,
		PubDate: t,
	}, nil
}

// 解析時間
func (s *BcciTV) parseTime(raw string) (time.Time, error) {
	suffixes := []string{"st", "nd", "rd", "th"}
	for _, s := range suffixes {
		raw = strings.ReplaceAll(raw, s, "")
	}
	layout := "2 Jan, 2006"
	return time.Parse(layout, raw)
}

func (s *BcciTV) FetchNewsDetail(url string, news *News) error {
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		return err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		return err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		return fmt.Errorf("建立分頁失敗: %w", err)
	}

	resp, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return err
	}
	log.Println(resp.Status())
	if resp.Status() != http.StatusOK {
		return fmt.Errorf("http Status is : %d", resp.Status())
	}

	// 因為新聞詳情中會有<table>屬性所以需要保留完整HTML，固不在進行篩選
	html, err := page.Locator("div.repor-bottom.mt-3").InnerHTML()
	if err != nil {
		return fmt.Errorf("取得 HTML 失敗: %w", err)
	}

	// 清理不需要的HTML屬性
	// 使用正則移除所有 style 屬性
	re := regexp.MustCompile(`\s*style="[^"]*"`)
	html = re.ReplaceAllString(html, "")
	// 使用正則移除所有 td 標籤中width 屬性
	re = regexp.MustCompile(`(<td[^>]*?)\s*width="[^"]*"`)
	html = re.ReplaceAllString(html, "$1")

	news.Content = strings.TrimSpace(html)

	return nil
}
