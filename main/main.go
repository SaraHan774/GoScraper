package main

import (
	"github.com/labstack/echo"
	"mainProject/scrapper"
	"os"
	"strings"
)


func handleHome(c echo.Context) error{
	//return c.String(http.StatusOK, "Hello, World!")
	return c.File("home.html")
}

func handleScrape(c echo.Context) error{
	defer os.Remove(fileName)
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment(fileName, fileName)
}

const fileName = "jobs.csv"

func main() {
	//scrapper.Scrape("python")
	e :=echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":8080"))
}