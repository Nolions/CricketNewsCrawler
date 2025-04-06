package playwright

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
	"log"
)

// NewPlaywright
// 初始化 Playwright
func NewPlaywright() (*playwright.Playwright, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("playwright 啟動失敗: %w", err)
	}

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
		return nil, fmt.Errorf("瀏覽器啟動失敗: %w", err)
	}

	return browser, nil
}

// NewPage
// 建立新分頁
func NewPage(browser playwright.Browser, hdrs map[string]string) (playwright.Page, error) {
	// 設置HTTP標頭，然後建立瀏覽器上下文
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		ExtraHttpHeaders: hdrs, // 設置 HTTP 標頭
	})
	if err != nil {
		log.Fatalf("建立上下文失敗: %v", err)
	}

	// 註冊捕捉請求的事件
	//context.OnRequest(func(request playwright.Request) {
	//	// 輸出請求的詳細信息
	//	fmt.Println("捕捉到請求:")
	//	fmt.Printf("請求 URL: %s\n", request.URL())
	//	fmt.Printf("請求方法: %s\n", request.Method())
	//	fmt.Printf("請求標頭: %v\n", request.Headers())
	//})

	page, err := context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("建立分頁失敗: %w", err)
	}

	return page, nil
}
