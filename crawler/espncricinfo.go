package crawler

import (
	"cricketNewsCrawler/helper"
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
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

	url := "https://www.espncricinfo.com/cricket-news"
	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status())
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
		return nil, fmt.Errorf("找不到新聞卡片: %w", err)
	}

	var newsList []News
	for i := 0; i < count; i++ {
		newsEl := newsEls.Nth(i)

		// 標題 <h2>
		titleLocator := newsEl.Locator("h2.ds-text-title-s.ds-font-bold.ds-text-typo")
		titleText, err := titleLocator.TextContent()
		if err != nil {
			continue
		}

		// 詳情連結 <a href="">
		linkLocator := newsEl.Locator("a")
		href, err := linkLocator.GetAttribute("href")
		if err != nil {
			continue
		}

		if strings.HasPrefix(href, "/") {
			href = s.Domain + href
		}

		// 時間 <span>
		timeLocator := newsEl.Locator("span.ds-text-compact-xs").First()
		timeText, err := timeLocator.TextContent()
		if err != nil {
			// TODO
		}

		t, err := s.parseTime(timeText)
		if err != nil {
			// TODO
		}

		// 描述（找沒有 class 的 <div>）
		descLocator := newsEl.Locator("div:not([class])").First()
		descText, err := descLocator.TextContent()
		if err != nil {
			// TODO
		}

		imgEl := newsEl.Locator("img")
		src, err := imgEl.GetAttribute("src")
		if err != nil {
			// TODO
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

	return newsList, nil
}

// 解析時間
func (s *ESPNCricinfo) parseTime(timeText string) (time.Time, error) {
	parts := strings.Split(timeText, "•")
	if len(parts) > 0 {
		date := strings.TrimSpace(parts[0])
		t, err := helper.CoverToTimestamp(date, "02-Jan-2006")
		if err != nil {
			log.Println("list: Failed to parse date", "date", date, "err", err)
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

	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		return err
	}
	log.Println(resp.Status())
	if resp.Status() != http.StatusOK {
		return fmt.Errorf("http Status is : %d", resp.Status())
	}

	locators := page.Locator(`p[class*="ds-text-comfortable-l"]`)

	count, err := locators.Count()
	if err != nil {
		return fmt.Errorf("新聞段落失敗: %w", err)
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

	return nil
}
