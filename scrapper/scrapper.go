package scrapper

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//start = 0 (first page)

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
	summary  string
}

//Scrape Indeed by a term
func Scrape(term string) {
	//https://kr.indeed.com ...
	var baseUrl string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"
	var jobs []extractedJob
	mainChannel := make(chan []extractedJob)

	totalPages := getPages(baseUrl)
	fmt.Println(totalPages)

	for i := 0; i < totalPages; i++ {
		go getPage(baseUrl, i, mainChannel) //returns array of jobs

	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-mainChannel
		jobs = append(jobs, extractedJobs...) //combine the result of all pages
	}

	writeJobs(jobs)
}

//csv 에 결과 작성
func writeJobs(jobs []extractedJob) {
	//create a file
	file, err := os.Create("jobs.csv")
	checkError(err)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"LINK", "TITLE", "LOCATION", "SALARY", "SUMMARY",
	}

	wErr := writer.Write(headers)
	checkError(wErr)

	for _, job := range jobs {
		jobSlice := []string{
			"https://kr.indeed.com/viewjob?jk=" + job.id,
			job.title,
			job.location,
			job.salary,
			job.summary,
		}
		jwErr := writer.Write(jobSlice)

		checkError(jwErr)
	}
}

func getPage(url string, page int, mainChannel chan<- []extractedJob) {

	var jobs []extractedJob
	channel := make(chan extractedJob)

	pageUrl := url + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting ", pageUrl)
	res, err := http.Get(pageUrl)
	checkError(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)
	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, s *goquery.Selection) {
		go extractJob(s, channel)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-channel
		jobs = append(jobs, job)
	}
	mainChannel <- jobs
}

func extractJob(card *goquery.Selection, channel chan<- extractedJob) {

	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())

	channel <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary,
	}
}

//Cleans a string
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {

	pages := 0

	res, err := http.Get(url)
	checkError(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkError(err)

	//div 의 클래스 이름이 아래인것을 찾는다.
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatal("request failed with status :", res.StatusCode)
	}
}
