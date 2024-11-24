package main

import (
	"encoding/csv"
	"fmt"
	"sync"

	// importing Colly
	"github.com/gocolly/colly"
	"log"
	"os"
)

type Product struct {
	Url, Image, Name, Price string
}

func main() {

	var products []Product

	// define a sync to filter visited URLs
	var visitedUrls sync.Map

	// instantiate a new collector object
	c := colly.NewCollector(
		colly.AllowedDomains("www.scrapingcourse.com"),
	)

	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

	// set up the proxy
	// https://free-proxy-list.net/
	err := c.SetProxy("http://38.54.59.154:80")
	if err != nil {
		log.Fatal("asd", err)
	}

	c.OnHTML("li.product", func(e *colly.HTMLElement) {

		// initialize a new Product instance
		product := Product{}

		// scrape the target data
		product.Url = e.ChildAttr("a", "href")
		product.Image = e.ChildAttr("img", "src")
		product.Name = e.ChildText(".product-name")
		product.Price = e.ChildText(".price")

		// add the product instance with scraped data to the list of products
		products = append(products, product)

	})

	// OnHTML callback for handling pagination
	c.OnHTML("a.next", func(e *colly.HTMLElement) {

		// extract the next page URL from the next button
		nextPage := e.Attr("href")

		// check if the nextPage URL has been visited
		if _, found := visitedUrls.Load(nextPage); !found {
			fmt.Println("scraping:", nextPage)
			// mark the URL as visited
			visitedUrls.Store(nextPage, struct{}{})
			// visit the next page
			err := e.Request.Visit(nextPage)
			if err != nil {
				fmt.Println("Error during Next visit", err)
			}
		}
	})

	// store the data to a CSV after extraction
	c.OnScraped(func(r *colly.Response) {

		// open the CSV file
		file, err := os.Create("products.csv")
		if err != nil {
			log.Fatalln("Failed to create output CSV file", err)
		}
		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				fmt.Println("Closing a file was unsuccessful", err)
			}
		}(file)

		// initialize a file writer
		writer := csv.NewWriter(file)

		// write the CSV headers
		headers := []string{
			"Url",
			"Image",
			"Name",
			"Price",
		}
		err = writer.Write(headers)
		if err != nil {
			fmt.Println("Failed to write csv headers", err)
		}

		// write each product as a CSV row
		for _, product := range products {
			// convert a Product to an array of strings
			record := []string{
				product.Url,
				product.Image,
				product.Name,
				product.Price,
			}

			// add a CSV record to the output file
			err = writer.Write(record)
			if err != nil {
				fmt.Println("Failed to write csv record", err)
			}
		}
		defer writer.Flush()
	})

	err = c.Visit("https://www.scrapingcourse.com/ecommerce")
	if err != nil {
		println("Error during Visit", err)
	}

}
