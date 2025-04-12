package playwright

import (
	"github.com/playwright-community/playwright-go"
	"github.com/rs/zerolog"
	"os"
)

var log = zerolog.New(os.Stdout).With().Timestamp().Logger()

// NewPlaywright
// 初始化 Playwright
func NewPlaywright() (*playwright.Playwright, error) {
	log.Info().Msg("Initializing Playwright")

	pw, err := playwright.Run()
	if err != nil {
		log.Error().Err(err).Msg("Failed to start Playwright")
		return nil, err
	}

	log.Info().Msg("Playwright started successfully")
	return pw, nil
}

// NewBrowser
// 啟動無頭瀏覽器
func NewBrowser(pw *playwright.Playwright) (playwright.Browser, error) {
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
		Args:     []string{"--disable-blink-features=AutomationControlled"},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to start browser")
		return nil, err
	}

	log.Info().Msg("Chromium browser started successfully")
	return browser, nil
}

// NewPage
// 建立新分頁
func NewPage(browser playwright.Browser, hdrs map[string]string) (playwright.Page, error) {
	log.Info().Msg("Creating new page")
	// 設置HTTP標頭，然後建立瀏覽器上下文
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		ExtraHttpHeaders: hdrs, // 設置 HTTP 標頭
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create context")
		return nil, err
	}

	// 註冊捕捉請求的事件
	context.OnRequest(func(request playwright.Request) {
		// 輸出請求的詳細信息
		//	log.Println("捕捉到請求:")
		//	log.Printf("請求 URL: %s\n", request.URL())
		//	log.Printf("請求方法: %s\n", request.Method())
		//	log.Printf("請求標頭: %v\n", request.Headers())
	})

	page, err := context.NewPage()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create page")
		return nil, err
	}

	log.Info().Msg("New page created successfully")
	return page, nil
}
