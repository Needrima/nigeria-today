package main

import (
	"fmt"
	"html/template"
	"log"

	"net/http"

	"github.com/gocolly/colly"
)

var tpl *template.Template

type news struct {
	Heading     string
	NewsLink    string
	PublishedAt string
}

type allnews struct {
	PunchNews, GuardianNews, VanguardNews, PremiumTimesNews, TheNationNews, AljazeeraNews []news
}

func init() {
	tpl = template.Must(template.ParseGlob("./templates/*"))
}

func main() {
	routes()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}

func CrawlNews(w http.ResponseWriter, r *http.Request) {
	collector := colly.NewCollector(
		colly.AllowedDomains(
		//https://thecable.ng/
		//https://dailypost.ng/headlines
		// 	"skysports.com",
		//www.completesports.com/
		),
	)

	collector.OnError(func(_ *colly.Response, err error) {
		log.Println("Collector error: ", err.Error())
		http.Error(w, "Something went wrong", 500)
	})

	collector.OnResponse(func(r *colly.Response) {
		log.Println("Visiting: %s\t StatusCode:", r.Request.URL, r.StatusCode)
	})

	// punch := getNews(".list-item article", ".entry-title a", ".entry-title a", ".entry-meta .meta-time span", "https://www.punchng.com", collector)
	// punch = filterNews(punch)

	// theGuardian := getNews(".row-3 .cell", "a .headline span", "a", "a .meta span", "https://www.guardian.ng", collector)
	// theGuardian = filterNews(guardian[3:])

	//theSun := getNews("article .jeg_postblock_content", "h3 a", "h3 a", ".jeg_post_meta .jeg_meta_date a", "https://www.sunnewsonline.com/", collector)
	//theSun = filterNews(theSun)

	//premiumTimes := getNews("article .jeg_postblock_content", "h3 a", "h3 a", ".jeg_post_meta .jeg_meta_date a", "https://www.premiumtimesng.com/", collector)
	//premiumTimes = filterNews(premiumTimes[1:])

	//aljazeera := getNews("article .gc__content", ".gc__header-wrap .gc__title a span", ".gc__header-wrap .gc__title a", ".gc__footer .gc__meta .gc__date .gc__date__date .date-simple", "https://www.aljazeera.com/where/nigeria/", collector)

	//saharaNews := getNews(".block-module-content", ".block-module-content-header span a", ".block-module-content-header span a", ".block-module-content-footer .block-module-content-footer-item-date", "https://www.saharareporters.com/", collector)
	//saharaNews = filterNews(saharaNews)

	// dailyTrust := getNews(".list_body__19fyx", "a", "a", ".list_category__1sVu4 span.list_time__1UhFn", "https://dailytrust.com", collector) // prefix media link with https://www.dailytrust.com
	// dailyTrust = dailyTrust[9:]

	thecable := getNews(".slider .flex-slider .flex-control-thumbs", " li a .flex-caption", "li a", "", "https://thecable.ng", collector)

	for _, v := range thecable {
		fmt.Println(v.Heading, ":\t", v.NewsLink)
		fmt.Println("--------------------------------------------------------------------------------------------")
	}

	tpl.ExecuteTemplate(w, "index.html", nil)
}

//helper functions

func getNews(htmlPath, headingElementHTMLPath, linkHtmlPath, publishedElementHTMLPath, URL string, collector *colly.Collector) []news {
	var News []news

	collector.OnHTML(htmlPath, func(ele *colly.HTMLElement) {
		heading := ele.ChildText(headingElementHTMLPath)
		mediaLink := ele.ChildAttr(linkHtmlPath, "href")
		published := ele.ChildText(publishedElementHTMLPath)

		Newnews := news{heading, mediaLink, published}
		News = append(News, Newnews)
	})

	collector.Visit(URL)

	return News
}

func filterNews(newsSlice []news) []news {
	present := map[string]bool{}

	var unique []news

	var count int

	for _, v := range newsSlice {
		if _, ok := present[v.Heading]; !ok {
			present[v.Heading] = true
			unique = append(unique, v)
		} else {
			count++
		}
	}

	fmt.Println("Recurring:", count)

	return unique
}

func routes() {
	http.HandleFunc("/", CrawlNews)

	http.Handle("/public/css/", http.StripPrefix("/public/css/", http.FileServer(http.Dir("public/css"))))
	http.Handle("/public/js/", http.StripPrefix("/public/js/", http.FileServer(http.Dir("public/js"))))
}
