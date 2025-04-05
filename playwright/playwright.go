package playwright

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
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
func NewPage(browser playwright.Browser) (playwright.Page, error) {
	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("建立分頁失敗: %w", err)
	}

	return page, nil
}
