package crawler

import (
	"fmt"
	"github.com/Nolions/CricketNewsCrawler/helper"
	pw "github.com/Nolions/CricketNewsCrawler/playwright"
	"github.com/playwright-community/playwright-go"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type SportSkeeda struct {
	Domain  string
	Headers map[string]string
}

func NewSportSkeeda() *SportSkeeda {
	return &SportSkeeda{
		Domain:  "https://www.sportskeeda.com",
		Headers: map[string]string{},
	}
}

// FetchNewsList
// 新聞列表
func (s *SportSkeeda) FetchNewsList() ([]News, error) {
	log.Info().Msg("Fetching news list from SportSkeeda")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		return nil, err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Playwright client")
		return nil, err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize browser")
		return nil, err
	}

	url := "https://www.sportskeeda.com/cricket"
	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		//Timeout:   playwright.Float(3000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create new page")
		return nil, err
	}

	log.Info().Int("status", resp.Status()).Msg("Page response status")
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
			log.Error().Err(err).Int("index", i).Msg("Failed to fetch title text for main news")
			return err
		}
		link, err := title.GetAttribute("href")
		if err != nil {
			log.Error().Err(err).Int("index", i).Msg("Failed to fetch news link for main news")
			return err
		}

		news.Title = titleText
		news.Link = s.Domain + link

		// 找 noscript 元素
		noscript := newsEl.Locator("noscript")
		html, err := noscript.InnerHTML()
		if err != nil {
			log.Error().Err(err).Int("index", i).Msg("Failed to fetch noscript HTML for main news")
			return err
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
		log.Error().Err(err).Msg("Failed to get count of secondary news elements")
		return fmt.Errorf("❌ 無法取得次新聞數量: %w", err)
	}

	for i := 0; i < count; i++ {
		news := News{}

		newsEl := locator.Nth(i)
		title := newsEl.Locator("a.feed-item-cta")
		titleText, err := title.TextContent()
		if err != nil {
			log.Error().Err(err).Int("index", i).Msg("Failed to fetch title text for secondary news")
			return err
		}
		news.Title = titleText

		link, err := title.GetAttribute("href")
		if err != nil || strings.TrimSpace(link) == "" {
			log.Error().Err(err).Int("index", i).Msg("Failed to fetch news link for secondary news")
			return err
		}
		news.Link = link

		noscript := newsEl.Locator("noscript")
		html, err := noscript.InnerHTML()
		if err != nil {
			log.Error().Err(err).Int("index", i).Msg("Failed to fetch noscript HTML for secondary news")
			return err
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
		log.Error().Err(err).Msg("Failed to initialize Playwright client")
		return err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize browser")
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
		//Timeout:   playwright.Float(3000),
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to go to URL")
		return err
	}

	log.Info().Int("status", resp.Status()).Msg("Page response status")
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
		log.Error().Err(err).Msg("Failed to fetch publication time")
		return err
	}

	t, err := helper.CoverToTimestamp(pubTime, time.RFC3339)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch publication time")
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
		log.Error().Err(err).Msg("Failed to get the count of paragraphs")
		return err
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		el := locators.Nth(i)

		// 要排除區塊
		skipParentCheck := el.Locator(`xpath=ancestor::div[contains(@class, 'post-author-info-parent') or contains(@class, 'bottom-tagline') or contains(@class, 'scrollable-content-holder')]`)
		parentCount, err := skipParentCheck.Count()
		if err != nil {
			log.Error().Int("index", i).Err(err).Msg("Failed to check parent elements")
			return err
		}
		if parentCount > 0 {
			log.Warn().Int("index", i).Msg("Skipped <p> due to parent exclusion")
			continue
		}

		text, err := el.TextContent()
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("Failed to fetch text content")
			continue
		}
		if strings.TrimSpace(text) == "" {
			log.Warn().Int("index", i).Msg("Skipped empty paragraph text")
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
