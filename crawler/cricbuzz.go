package crawler

import (
	"fmt"
	pw "github.com/Nolions/CricketNewsCrawler/playwright"
	"github.com/playwright-community/playwright-go"
	"html"
	"net/http"
	"strings"
	"time"
)

type Cricbuzz struct {
	Domain  string
	Headers map[string]string
}

func NewCricbuzz() *Cricbuzz {
	headers := map[string]string{
		"sec-ch-ua":                 `"Chromium";v="133", "Not(A:Brand";v="99"`,
		"sec-ch-ua-mobile":          "?0",
		"sec-ch-ua-platform":        `"macOS"`,
		"upgrade-insecure-requests": "1",
		"user-agent":                "lla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	}

	return &Cricbuzz{
		Domain:  "https://www.cricbuzz.com",
		Headers: headers,
	}
}

// FetchNewsList
// 新聞列表
func (c *Cricbuzz) FetchNewsList() ([]News, error) {
	log.Info().Msg("Starting FetchNewsList")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize Playwright")
		return nil, err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize Playwright")
		return nil, err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, c.Headers)
	if err != nil {
		log.Error().Err(err).Msg("failed to create new page")
		return nil, err
	}

	url := "https://www.cricbuzz.com/cricket-news"
	// 打開網頁
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(60000),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to navigate to news list page")
		return nil, err
	}
	log.Println(resp.Status())
	if resp.Status() != http.StatusOK {
		return nil, fmt.Errorf("http Status is : %d", resp.Status())
	}

	newsLocator := page.Locator(".cb-nws-intr")
	err = newsLocator.First().WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(20000),
	})

	if err != nil {
		log.Error().Err(err).Str("url", url).Msgf("initial news items not found: %v", err)
		return nil, err
	}

	newsLocator = page.Locator(".cb-nws-intr")
	err = newsLocator.First().WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(20000),
	})
	if err != nil {
		log.Error().Err(err).Str("url", url).Msgf("initial news items not found: %v", err)
		return nil, err
	}

	newsList, err := c.extractNewsList(page)
	if err != nil {
		log.Error().Err(err).Msg("failed to extract news list")
		return nil, err
	}
	for i := 0; i < 10; i++ {
		if _, err := page.Evaluate(`window.scrollBy(0, window.innerHeight);`); err != nil {
			return nil, err
		}
		time.Sleep(500 * time.Millisecond)
	}

	// 提取封面圖片
	err = c.extractCoverImages(page, &newsList)
	if err != nil {
		log.Error().Err(err).Msg("failed to extract cover images")
		return nil, err
	}

	return newsList, nil
}

func (c *Cricbuzz) extractNewsList(page playwright.Page) ([]News, error) {
	var newsList []News

	// 找到所有的新聞項目
	docEl := page.Locator("#news-list").First()

	newsElements := docEl.Locator(".cb-lst-itm")
	count, err := newsElements.Count()
	if err != nil {
		log.Error().Err(err).Msg("failed to extract cover images")
		return nil, err
	}

	for i := 0; i < count; i++ {
		// 標題、連結
		newsElement := newsElements.Nth(i)
		title := newsElement.Locator("a.text-hvr-underline")
		titleText, err := title.TextContent()
		if err != nil {
			log.Warn().Int("index", i).Err(err).Msg("failed to get title text")
			continue
		}

		link, err := title.GetAttribute("href")
		if err != nil {
			log.Warn().Int("index", i).Err(err).Msg("failed to get link attribute")
			continue
		}

		// 取得描述
		desc := newsElement.Locator(".cb-nws-intr")
		descText, err := desc.TextContent()
		if err != nil {
			log.Warn().Int("index", i).Err(err).Msg("failed to get description text")
			continue
		}

		// 將每條新聞加入到新聞列表中
		newsList = append(newsList, News{
			Title: titleText,
			Link:  c.Domain + link,
			Desc:  descText,
		})
	}

	return newsList, nil
}

func (c *Cricbuzz) extractCoverImages(page playwright.Page, newsList *[]News) error {
	// 找到所有的 meta 標籤，並提取封面圖片
	metaTags := page.Locator("meta[itemprop='url']")
	count, err := metaTags.Count()
	if err != nil {
		log.Error().Err(err).Msg("failed to count meta tags")
		return err
	}

	log.Info().Int("count", count).Msg("meta tags found for cover images")
	for i := 0; i < count; i++ {
		if i >= len(*newsList) {
			break
		}

		contentValue, err := metaTags.Nth(i).GetAttribute("content")
		if err != nil {
			log.Warn().Int("index", i).Err(err).Msg("failed to get content attribute")
			continue
		}

		// 更新新聞項目的封面圖片
		(*newsList)[i].Cover = contentValue
	}

	return nil
}

// FetchNewsDetail
// 取得內文
func (c *Cricbuzz) FetchNewsDetail(url string, news *News) error {
	url = strings.TrimSpace(url)
	log.Info().Str("url", url).Msg("Starting FetchNewsDetail")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		return err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize Playwright")
		return err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, c.Headers)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to navigate to news detail page")
		return err
	}

	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(60000),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to navigate to news detail page")
		return err
	}

	log.Info().Int("status", resp.Status()).Str("url", url).Msg("detail page loaded")
	if resp.Status() != http.StatusOK {
		return fmt.Errorf("http Status is : %d", resp.Status())
	}

	for i := 0; i < 10; i++ {
		page.Evaluate("window.scrollBy(0, document.body.scrollHeight)")
		//page.WaitForTimeout(1000)
	}

	paras := page.Locator("p.cb-nws-para:not(:has(b))")
	count, err := paras.Count()
	if err != nil {
		log.Error().Err(err).Msg("failed to count paragraph elements")
		return err
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		text, err := paras.Nth(i).TextContent()
		if err != nil {
			log.Warn().Int("index", i).Err(err).Msg("failed to get paragraph text")
			continue
		}

		//  跳過空字串
		if text == "" {
			continue
		}
		builder.WriteString("<p>")
		builder.WriteString(html.EscapeString(text))
		builder.WriteString("</p>")
	}

	news.Title = builder.String()

	return nil
}
