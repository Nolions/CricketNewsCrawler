package crawler

import (
	"cricketNewsCrawler/helper"
	pw "cricketNewsCrawler/playwright"
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
	"net/http"
	"strings"
)

type Ndtv struct {
	Domain  string
	Headers map[string]string
}

func NewNdTv() *Ndtv {
	return &Ndtv{
		Domain:  "https://www.bcci.tv",
		Headers: map[string]string{},
	}
}

// FetchNewsList
// 新聞列表
func (s *Ndtv) FetchNewsList() ([]News, error) {
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
		return nil, err
	}

	url := "https://sports.ndtv.com/cricket/news"
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

	locators := page.Locator("ul#container_listing >> div.lst-pg-a")
	count, err := locators.Count()
	if err != nil {
		return nil, err
	}

	var newsList []News
	for i := 0; i < count; i++ {
		locator := locators.Nth(i)
		titleEL := locator.Locator("a.lst-pg_ttl")
		title, err := titleEL.TextContent()
		if err != nil {
			// TODO
			continue
		}

		link, err := titleEL.GetAttribute("href")
		if err != nil {
			// TODO
			continue
		}

		descEl := locator.Locator("p.lst-pg_txt.txt_tct.txt_tct-three")
		desc, err := descEl.TextContent()
		if err != nil {
			// TODO
		}

		imgEl := locator.Locator("img.lz_img.crd_img-full")
		src, err := imgEl.GetAttribute("src")
		if err != nil {
			// TODO
		}

		pubTimeEl := locator.Locator("span.lst-a_pst_lnk").First()
		pubTime, err := pubTimeEl.TextContent()
		if err != nil {
			// TODO
		}

		t, err := helper.CoverToTimestamp(pubTime, "Jan 2, 2006")

		newsList = append(newsList, News{
			Title:   title,
			Link:    s.Domain + link,
			Desc:    desc,
			Cover:   src,
			PubDate: t,
		})
	}

	return newsList, nil
}

// FetchNewsDetail
// 取得內文
func (s *Ndtv) FetchNewsDetail(url string, news *News) error {
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
		return err
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

	locators := page.Locator("div.story__content").First().Locator("p")
	count, err := locators.Count()
	if err != nil {
		return err
	}

	var builder strings.Builder
	for i := 0; i < count; i++ {
		locator := locators.Nth(i)
		content, err := locator.TextContent()
		if err != nil {
			// TODO
			continue
		}

		if content == "" {
			continue
		}

		builder.WriteString("<p>")
		builder.WriteString(content)
		builder.WriteString("</p>")
	}

	news.Content = builder.String()

	return nil
}
