package crawler

import (
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type SportSkeeda struct {
	Domain string
}

func NewSportSkeeda() *SportSkeeda {
	return &SportSkeeda{
		Domain: "https://www.sportskeeda.com",
	}
}

func (s *SportSkeeda) FetchNewsList() ([]News, error) {
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

	page, err := pw.NewPage(browser)
	if err != nil {
		return nil, fmt.Errorf("建立分頁失敗: %w", err)
	}

	url := "https://www.sportskeeda.com/cricket"
	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		//Timeout:   playwright.Float(3000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return nil, err
	}
	log.Println(resp.Status())
	if resp.Status() != http.StatusOK {
		return nil, fmt.Errorf("http Status is : %d", resp.Status())
	}

	var newsList []News
	mainNewsEls := page.Locator("div.sport-feed-item-primary")
	s.extractMainNewsList(mainNewsEls, &newsList)
	secNewsEls := page.Locator("div.sports-feed-item-secondary-element")
	s.extractSecondsNewsList(secNewsEls, &newsList)

	return newsList, nil
}

func (s *SportSkeeda) extractMainNewsList(locator playwright.Locator, newsList *[]News) error {
	count, err := locator.Count()
	if err != nil {
		return fmt.Errorf("❌ 無法取得次新聞數量: %w", err)
	}

	for i := 0; i < count; i++ {
		news := News{}
		newsEl := locator.Nth(i)

		title := newsEl.Locator("a.feed-item-cta")
		titleText, err := title.TextContent()
		if err != nil {
			return fmt.Errorf("⚠️ 第 %d 條新聞標題抓取失敗或為空: %v", i, err)
		}
		link, err := title.GetAttribute("href")
		if err != nil {
			return fmt.Errorf("⚠️ 第 %d 條新聞連結抓取失敗或為空: %v", i, err)
		}

		news.Title = titleText
		news.Link = s.Domain + link

		// 找 noscript 元素
		noscript := newsEl.Locator("noscript")
		html, err := noscript.InnerHTML()
		if err != nil {
			return fmt.Errorf("第 %d 條 noscript 取得失敗: %v", i, err)
		}

		re := regexp.MustCompile(`<img[^>]+alt="([^"]+)"[^>]+src="([^"]+)"`)
		match := re.FindStringSubmatch(html)
		if len(match) >= 3 {
			news.Desc = match[2]
		} else {
			log.Printf("⚠️ 第 %d 條 noscript 中找不到有效的 <img> 標籤", i)
			continue
		}

		*newsList = append(*newsList, news)
	}

	return nil
}

func (s *SportSkeeda) extractSecondsNewsList(locator playwright.Locator, newsList *[]News) error {
	count, err := locator.Count()
	if err != nil {
		log.Fatalf("獲取次新聞數量失敗: %v", err)
		return fmt.Errorf("❌ 無法取得次新聞數量: %w", err)
	}

	for i := 0; i < count; i++ {
		news := News{}

		newsEl := locator.Nth(i)
		title := newsEl.Locator("a.feed-item-cta")
		titleText, err := title.TextContent()
		if err != nil {
			return fmt.Errorf("❌ 第 %d 條新聞標題抓取失敗或為空: %w", i, err)
		}
		news.Title = titleText

		link, err := title.GetAttribute("href")
		if err != nil || strings.TrimSpace(link) == "" {
			return fmt.Errorf("❌ 第 %d 條新聞連結抓取失敗或為空: %w", i, err)
		}
		news.Link = link

		noscript := newsEl.Locator("noscript")
		html, err := noscript.InnerHTML()
		if err != nil {
			return fmt.Errorf("❌ 第 %d 條新聞 noscript 抓取失敗: %w", i, err)
		}
		re := regexp.MustCompile(`<img[^>]+src="([^"]+)"`)
		match := re.FindStringSubmatch(html)
		if len(match) >= 2 {
			news.Cover = match[1]
		}

		*newsList = append(*newsList, news)
	}

	return nil
}

func (s *SportSkeeda) FetchNewsDetail(url string) (string, error) {
	return "", nil
}
