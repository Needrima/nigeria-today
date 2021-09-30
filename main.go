package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gocolly/colly"
)

var tpl *template.Template

type covidInfo struct {
	Country      string `json:"country"`
	TotalCases   int    `json:"cases"`
	TodaysCases  int    `json:"todayCases"`
	TotalDeaths  int    `json:"deaths"`
	TodaysDeaths int    `json:"todayDeaths"`
	Recovered    int    `json:"recovered"`
	Active       int    `json:"active"`
	Critical     int    `json:"critical"`
	CPM          int    `json:"casesPerOneMillion"`  //Cases per million
	DPM          int    `json:"deathsPerOneMillion"` //Deaths per million
	TotalTests   int    `json:"totalTests"`
	TPM          int    `json:"testsPerOneMillion"` //Tests per million
}

type news struct {
	Heading     string
	NewsLink    string
	PublishedAt string
}

type allNews struct {
	PunchNews, GuardianNews, SunNews, PremiumTimesNews, AlJazeeraNews, SaharaNews, DailyTrustNews, DailyPostNews, SkySportsNews, CompleteSportsNews1, CompleteSportsNews2 []news
	covidInfo
}

func init() {
	tpl = template.Must(template.ParseGlob("./templates/*"))
}

func main() {
	routes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.ListenAndServe(":"+port, nil)
}

func CrawlNews(w http.ResponseWriter, r *http.Request) {
	collector := colly.NewCollector()

	collector.OnError(func(_ *colly.Response, err error) {
		log.Println("Collector error: ", err.Error())
		http.Error(w, "Something went wrong", 500)
		return
	})

	collector.OnResponse(func(r *colly.Response) {
		log.Println("Visiting: %s\t StatusCode:", r.Request.URL, r.StatusCode)
	})

	punch := getNews(".list-item article", ".entry-title a", ".entry-title a", ".entry-meta .meta-time span", "https://www.punchng.com", collector)
	punch = filterNews(punch)

	theGuardian := getNews(".row-3 .cell", "a .headline span", "a", "a .meta span", "https://www.guardian.ng", collector)
	theGuardian = filterNews(theGuardian)

	theSun := getNews("article .jeg_postblock_content", "h3 a", "h3 a", ".jeg_post_meta .jeg_meta_date a", "https://www.sunnewsonline.com/", collector)
	theSun = filterNews(theSun)

	premiumTimes := getNews("article .jeg_postblock_content", "h3 a", "h3 a", ".jeg_post_meta .jeg_meta_date a", "https://www.premiumtimesng.com/", collector)
	premiumTimes = filterNews(premiumTimes)

	aljazeera := getNews("article .gc__content", ".gc__header-wrap .gc__title a span", ".gc__header-wrap .gc__title a", ".gc__footer .gc__meta .gc__date .gc__date__date .date-simple", "https://www.aljazeera.com/where/nigeria/", collector)

	saharaNews := getNews(".block-module-content", ".block-module-content-header span a", ".block-module-content-header span a", ".block-module-content-footer .block-module-content-footer-item-date", "https://www.saharareporters.com/", collector)
	saharaNews = filterNews(saharaNews)

	dailyTrust := getNews(".list_body__19fyx", "a", "a", ".list_category__1sVu4 span.list_time__1UhFn", "https://dailytrust.com", collector) // prefix media link with https://www.dailytrust.com
	dailyTrust = dailyTrust[9:]

	dailypost := getNews(".mvp-blog-story-wrap", "a .mvp-blog-story-in .mvp-blog-story-text h2", "a", "a .mvp-blog-story-in .mvp-blog-story-text .mvp-cat-date-wrap .mvp-cd-date", "https://dailypost.ng/headlines/", collector)

	skysports := getNews(".sdc-site-tile__body-main", ".sdc-site-tile__headline a span", ".sdc-site-tile__headline a", ".sdc-site-tile__info .sdc-site-tile__tag a ", "https://www.skysports.com/", collector) //prefix link with https://www.skysports.com/

	completesports := getNews(".td", ".item-title a span", ".item-title a", ".meta-item-date a span", "https://www.completesports.com/", collector)
	completesports2 := getNews(".item-sub", ".item-title a", ".item-title a", ".meta-items .meta-item-date span", "https://www.completesports.com/", collector)
	completesports2 = filterNews(completesports2)

	c, _ := getCovidInfo("niGeria")

	news := allNews{punch, theGuardian, theSun, premiumTimes, aljazeera, saharaNews, dailyTrust, dailypost, skysports, completesports, completesports2, c}

	//fmt.Println(c)

	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "index.html", news)
	} else if r.Method == http.MethodPost {
		countryName := r.FormValue("country")

		c, err := getCovidInfo(countryName)

		if err != nil {
			if err.Error() == "Country not found" {
				http.Error(w, err.Error(), 400)
				return
			}

			http.Error(w, err.Error(), 500)
			return
		}

		news.covidInfo = c
		tpl.ExecuteTemplate(w, "index.html", news)
	}
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

// func foo(w http.ResponseWriter, r *http.Request) {
// 	tpl.ExecuteTemplate(w, "index.html", nil)
// }

func routes() {
	http.HandleFunc("/", CrawlNews)

	http.Handle("/public/css/", http.StripPrefix("/public/css/", http.FileServer(http.Dir("public/css"))))
	http.Handle("/public/js/", http.StripPrefix("/public/js/", http.FileServer(http.Dir("public/js"))))
}

func getCovidInfo(name string) (covidInfo, error) {
	url := fmt.Sprintf("https://coronavirus-19-api.herokuapp.com/countries/%s", name)
	r, err := http.Get(url)
	if err != nil {
		return covidInfo{}, errors.New("Something went wrong")
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return covidInfo{}, errors.New("Something went wrong")
	}

	if content := string(b); content == "Country not found" {
		return covidInfo{}, errors.New(content)
	}

	var c covidInfo

	err = json.Unmarshal(b, &c)
	if err != nil {
		return covidInfo{}, errors.New("Something went wrong")
	}

	return c, nil
}
