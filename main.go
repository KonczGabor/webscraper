package main

import (
	"encoding/csv"
	"fmt"

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

	// instantiate a new collector object
	c := colly.NewCollector(
		colly.AllowedDomains("www.scrapingcourse.com"),
	)

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

	err := c.Visit("https://www.scrapingcourse.com/ecommerce")
	if err != nil {
		println("Error during Visit", err)
	}

}
