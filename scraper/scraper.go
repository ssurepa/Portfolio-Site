package scraper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	jobId       string
	jobTitle    string
	jobCompany  string
	jobLocation string
	jobSalary   string
}

//Scrape Indeed
func Scrape(term string) {
	var baseURL string = "https://www.indeed.com/jobs?q=" + term + "&l=07670&limit=50&radius=25"
	var jobs []extractedJob
	c_getPage := make(chan []extractedJob)
	totalPages := getLastPage(baseURL)

	for i := 0; i < totalPages; i++ {
		go getPage(i, baseURL, c_getPage)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c_getPage
		jobs = append(jobs, extractedJobs...)
	}
	writeJobs(jobs)
	fmt.Println("Done. Extracted", len(jobs), "jobs!")
}

//writeJobs save search result in csv
func writeJobs(jobs []extractedJob) {
	file, err := os.Create("go_jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush()

	headers := []string{"ID", "Title", "Company", "Location", "Salary"}
	wErr := w.Write(headers)
	checkErr(wErr)

	c_writeCSV := make(chan []string)

	for _, job := range jobs {
		go writeCSV(job, c_writeCSV)
	}

	for i := 0; i < len(jobs); i++ {
		jwErr := w.Write(<-c_writeCSV)
		checkErr(jwErr)
	}

}

func writeCSV(job extractedJob, c_writeCSV chan<- []string) {
	c_writeCSV <- []string{
		"https://www.indeed.com/viewjob?jk=" + job.jobId,
		job.jobTitle,
		job.jobCompany,
		job.jobLocation,
		job.jobSalary,
	}
}

//getPage visit each page
func getPage(page int, url string, c_getPage chan<- []extractedJob) {
	var jobs []extractedJob
	c_extractedJob := make(chan extractedJob)
	pageURL := url + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting", pageURL)
	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".tapItem")
	searchCards.Each(func(i int, jobCard *goquery.Selection) {
		go extractJob(jobCard, c_extractedJob)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c_extractedJob
		jobs = append(jobs, job)
	}

	c_getPage <- jobs
}

//extractJob extract job attributes
func extractJob(jobCard *goquery.Selection, c_extractedJob chan<- extractedJob) {
	jobId, _ := jobCard.Attr("data-jk")
	jobTitle := CleanString(jobCard.Find(".jobTitle").Text())
	jobCompany := CleanString(jobCard.Find(".companyName").Text())
	jobLocation := CleanString(jobCard.Find(".companyLocation").Text())
	jobSalary := jobCard.Find(".metadata").Text()
	c_extractedJob <- extractedJob{
		jobId:       jobId,
		jobTitle:    jobTitle,
		jobCompany:  jobCompany,
		jobLocation: jobLocation,
		jobSalary:   jobSalary}

}

// getLastPage get last page that needs to be searched
func getLastPage(url string) int {
	lastPage := 0

	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		lastPage = s.Find("a").Length()
	})

	return lastPage
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request Failed with Status:", res.StatusCode)
	}
}

//CleanString
func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
