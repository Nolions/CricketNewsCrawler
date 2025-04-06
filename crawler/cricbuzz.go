package crawler

import (
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"html"
	"log"
	"net/http"
	"strings"
)

type Cricbuzz struct {
	Domain  string
	Headers map[string]string
}

func NewCricbuzz() *Cricbuzz {
	return &Cricbuzz{
		Domain:  "https://www.cricbuzz.com",
		Headers: map[string]string{},
	}
}

func (c *Cricbuzz) FetchNewsList() ([]News, error) {
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

	page, err := pw.NewPage(browser, c.Headers)
	if err != nil {
		return nil, fmt.Errorf("建立分頁失敗: %w", err)
	}

	url := "https://www.cricbuzz.com/cricket-news"
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

	newsList, err := c.extractNewsList(page)
	if err != nil {
		return nil, fmt.Errorf("提取新聞列表失敗: %w", err)
	}

	// 提取封面圖片
	err = c.extractCoverImages(page, &newsList)
	if err != nil {
		return nil, fmt.Errorf("提取封面圖片失敗: %w", err)
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
		log.Fatalf("獲取新聞數量失敗: %v", err)
	}

	for i := 0; i < count; i++ {
		// 標題、連結
		newsElement := newsElements.Nth(i)
		title := newsElement.Locator("a.text-hvr-underline")
		titleText, err := title.TextContent()
		if err != nil {
			continue
		}

		link, err := title.GetAttribute("href")
		if err != nil {
			continue
		}

		// 取得描述
		desc := newsElement.Locator(".cb-nws-intr")
		descText, err := desc.TextContent()
		if err != nil {
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
		log.Fatalf("獲取新聞數量失敗: %v", err)
	}
	for i := 0; i < count; i++ {
		if i >= len(*newsList) {
			break
		}

		contentValue, err := metaTags.Nth(i).GetAttribute("content")
		if err != nil {
			log.Printf("獲取第 %d 個 meta 標籤的 content 失敗: %v", i, err)
			continue
		}

		// 更新新聞項目的封面圖片
		(*newsList)[i].Cover = contentValue
	}

	return nil
}

func (c *Cricbuzz) FetchNewsDetail(url string, news *News) error {
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

	page, err := pw.NewPage(browser, c.Headers)
	if err != nil {
		return fmt.Errorf("建立分頁失敗: %w", err)
	}

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

	paras := page.Locator("p.cb-nws-para:not(:has(b))")
	count, err := paras.Count()
	if err != nil {
		return err
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		text, err := paras.Nth(i).TextContent()
		if err != nil {
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
