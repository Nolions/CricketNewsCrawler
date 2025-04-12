package crawler

import (
	"cricketNewsCrawler/helper"
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"net/http"
	"strings"
	"time"
)

type ESPNCricinfo struct {
	Domain  string
	Headers map[string]string
}

func NewESPNCricinfo() *ESPNCricinfo {
	headers := map[string]string{
		"sec-ch-ua":                 `"Chromium";v="133", "Not(A:Brand";v="99"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"macOS"`,
		"upgrade-insecure-requests": "1",
		"user-agent":                "lla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	return &ESPNCricinfo{
		Domain:  "https://www.espncricinfo.com",
		Headers: headers,
	}
}

// FetchNewsList
// 新聞列表
func (s *ESPNCricinfo) FetchNewsList() ([]News, error) {
	log.Info().Msg("Fetching news list from ESPN Cricinfo")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Playwright")
		return nil, err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start browser")
		return nil, err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new page")
		return nil, err
	}

	url := "https://www.espncricinfo.com/cricket-news"
	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to navigate to the URL")
		return nil, err
	}

	log.Info().Int("status", resp.Status()).Msg("Page response status")
	if resp.Status() != http.StatusOK {
		return nil, fmt.Errorf("http Status is : %d", resp.Status())
	}

	for i := 0; i < 10; i++ {
		page.Evaluate("window.scrollBy(0, document.body.scrollHeight)")
		//page.WaitForTimeout(1000)
	}

	newsEls := page.Locator("div.ds-border-b.ds-border-line.ds-p-4")
	count, err := newsEls.Count()
	if err != nil {
		log.Error().Err(err).Msg("Failed to find news cards")
		return nil, err
	}

	var newsList []News
	for i := 0; i < count; i++ {
		newsEl := newsEls.Nth(i)

		// 標題 <h2>
		titleLocator := newsEl.Locator("h2.ds-text-title-s.ds-font-bold.ds-text-typo")
		titleText, err := titleLocator.TextContent()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get title text")
			continue
		}

		// 詳情連結 <a href="">
		linkLocator := newsEl.Locator("a")
		href, err := linkLocator.GetAttribute("href")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get link attribute")
			continue
		}

		if strings.HasPrefix(href, "/") {
			href = s.Domain + href
		}

		// 時間 <span>
		timeLocator := newsEl.Locator("span.ds-text-compact-xs").First()
		timeText, err := timeLocator.TextContent()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get time text")
		}

		t, err := s.parseTime(timeText)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse time")
		}

		// 描述（找沒有 class 的 <div>）
		descLocator := newsEl.Locator("div:not([class])").First()
		descText, err := descLocator.TextContent()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get description text")
		}

		imgEl := newsEl.Locator("img")
		src, err := imgEl.GetAttribute("src")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get image source")
		}
		cover := ""
		if !strings.Contains(src, "lazyimage-noaspect.svg") {
			cover = src
		}

		newsList = append(newsList, News{
			Title:   titleText,
			Desc:    descText,
			Link:    href,
			PubDate: t,
			Cover:   cover,
		})
	}

	log.Info().Int("news_count", len(newsList)).Msg("Fetched news list successfully")
	return newsList, nil
}

// 解析時間
func (s *ESPNCricinfo) parseTime(timeText string) (time.Time, error) {
	log.Info().Str("timeText", timeText).Msg("parse Time")
	parts := strings.Split(timeText, "•")
	if len(parts) > 0 {
		date := strings.TrimSpace(parts[0])
		t, err := helper.CoverToTimestamp(date, "02-Jan-2006")
		if err != nil {
			log.Error().Err(err).Str("date", date).Msg("Failed to parse date")
			return time.Time{}, err
		} else {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("無法取得新聞發布時間")
}

// FetchNewsDetail
// 取得內文
func (s *ESPNCricinfo) FetchNewsDetail(url string, news *News) error {
	log.Info().Str("url", url).Msg("Fetching news detail")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Playwright")
		return err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start browser")
		return err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new page")
		return err
	}

	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to navigate to the URL")
		return err
	}

	if resp.Status() != http.StatusOK {
		return fmt.Errorf("http Status is : %d", resp.Status())
	}

	locators := page.Locator(`p[class*="ds-text-comfortable-l"]`)

	count, err := locators.Count()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get paragraph text")
		return err
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		locator := locators.Nth(i).Locator("div")
		text, err := locator.TextContent()
		if err != nil {
			// TODO
			continue
		}
		if text == "" {
			continue
		}

		builder.WriteString("<p>")
		builder.WriteString(text)
		builder.WriteString("</p>")
	}
	news.Content = builder.String()

	log.Info().Str("url", url).Msg("News detail fetched successfully")
	return nil
}
