package crawler

import (
	"cricketNewsCrawler/helper"
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
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

func (s *SportSkeeda) FetchNewsDetail(url string, news *News) error {
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

	page, err := pw.NewPage(browser)
	if err != nil {
		return fmt.Errorf("建立分頁失敗: %w", err)
	}

	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		//Timeout:   playwright.Float(3000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		return err
	}

	log.Println(resp.Status())
	if resp.Status() != http.StatusOK {
		return fmt.Errorf("http Status is : %d", resp.Status())
	}

	// 發布時間
	err = s.extractNewsPubTime(page, news)
	if err != nil {
		// TODO ERROR
		//return nil
	}

	// 內文
	err = s.extractNewsContent(page, news)
	if err != nil {
		// TODO ERROR
		//return nil
	}

	return nil
}

// 取得發布時間
func (s *SportSkeeda) extractNewsPubTime(page playwright.Page, news *News) error {
	pubTimeEl := page.Locator("div.date-pub").First()
	pubTime, err := pubTimeEl.GetAttribute("data-iso-string")
	if err != nil {
		return err
	}

	t, err := helper.CoverToTimestamp(pubTime, time.RFC3339)
	if err != nil {
		return err
	}
	news.PubDate = t

	return nil
}

// 取得內文
func (s *SportSkeeda) extractNewsContent(page playwright.Page, news *News) error {
	locators := page.Locator("p[data-idx]")
	count, err := locators.Count()
	if err != nil {
		return fmt.Errorf("獲取 <p data-idx> 數量失敗: %w", err)
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		el := locators.Nth(i)

		// 要排除區塊
		skipParentCheck := el.Locator(`xpath=ancestor::div[contains(@class, 'post-author-info-parent') or contains(@class, 'bottom-tagline') or contains(@class, 'scrollable-content-holder')]`)
		parentCount, err := skipParentCheck.Count()
		if err != nil {
			return fmt.Errorf("檢查父層失敗: %w", err)
		}
		if parentCount > 0 {
			continue
		}

		text, err := el.TextContent()
		if err != nil {
			log.Printf("⚠️ 第 %d 個 <p> 抓取失敗: %v", i, err)
			continue
		}
		if strings.TrimSpace(text) == "" {
			continue
			//result = append(result, strings.TrimSpace(text))
		}

		if text == "" {
			continue
		}
		builder.WriteString("<p>")
		builder.WriteString(strings.TrimSpace(text))
		builder.WriteString("</p>")
	}

	news.Content = builder.String()

	return nil
}
