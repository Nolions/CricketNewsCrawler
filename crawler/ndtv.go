package crawler

import (
	"fmt"
	"github.com/Nolions/CricketNewsCrawler/helper"
	pw "github.com/Nolions/CricketNewsCrawler/playwright"
	"github.com/playwright-community/playwright-go"
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
	log.Info().Msg("Starting FetchNewsList")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Playwright")
		return nil, err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to launch browser")
		return nil, err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create page")
		return nil, err
	}

	url := "https://sports.ndtv.com/cricket/news"
	resp, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to load page")
		return nil, err
	}

	log.Info().Int("status", resp.Status()).Msg("Page response status")
	if resp.Status() != http.StatusOK {
		return nil, fmt.Errorf("http Status is : %d", resp.Status())
	}

	locators := page.Locator("ul#container_listing >> div.lst-pg-a")
	count, err := locators.Count()
	if err != nil {
		log.Error().Err(err).Msg("Failed to count article blocks")
		return nil, err
	}

	var newsList []News
	for i := 0; i < count; i++ {
		locator := locators.Nth(i)
		titleEL := locator.Locator("a.lst-pg_ttl")
		title, err := titleEL.TextContent()
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("Failed to get title")
			continue
		}

		link, err := titleEL.GetAttribute("href")
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("Failed to get link")
			continue
		}

		descEl := locator.Locator("p.lst-pg_txt.txt_tct.txt_tct-three")
		desc, err := descEl.TextContent()
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("Failed to get description")
		}

		imgEl := locator.Locator("img.lz_img.crd_img-full")
		src, err := imgEl.GetAttribute("src")
		if err != nil {
			log.Warn().Err(err).Int("index", i).Msg("Failed to get href")
		}

		pubTimeEl := locator.Locator("span.lst-a_pst_lnk").First()
		pubTime, err := pubTimeEl.TextContent()
		if err != nil {
			log.Warn().Err(err).Str("pubTime", pubTime).Msg("Failed to parse publish time")
		}

		t, err := helper.CoverToTimestamp(pubTime, "Jan 2, 2006")
		if err != nil {
			log.Warn().Err(err).Str("pubTime", pubTime).Msg("Failed to parse publish time")
		}

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
	log.Info().Str("url", url).Msg("Starting FetchNewsDetail")
	pwClient, err := pw.NewPlaywright()
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize Playwright")
		return err
	}
	defer pwClient.Stop()

	browser, err := pw.NewBrowser(pwClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to launch browser")
		return err
	}
	defer browser.Close()

	page, err := pw.NewPage(browser, s.Headers)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create page")
		return err
	}

	resp, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to navigate to article page")
		return err
	}
	log.Info().Int("status", resp.Status()).Msg("Detail page response status")

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
			log.Warn().Err(err).Int("index", i).Msg("Failed to read paragraph")
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
