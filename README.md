# CricketNewsCrawler

爬蟲練習，目標是以各種模擬瀏覽器行為套件(EX:playwright、chromedp或selenium)，並最好以無頭瀏覽器的方式去爬取資料。

預計要爬取頁面與內文如下：
1. 從新聞列表頁面中那道新聞標題、介紹(描述)、封面與時間等資訊
2. 從新聞詳情頁面中到內文，其中內文只需要拿到文字內文，圖片或表格等資訊都不需要，另外依照內文分段用<p></p>進行分段處理。

然後初步只針對第一是時間載入的頁面進行爬取，部分需要透過點擊行為去觸發載入更過資訊等後續再行研究。

## Other

### 目標網站

- [x] https://www.cricbuzz.com/cricket-news
- [ ] https://www.espncricinfo.com/cricket-news
- [ ] https://sports.ndtv.com/cricket/news
- [x] https://www.sportskeeda.com/cricket
- [ ] https://www.bcci.tv/international/men/news
