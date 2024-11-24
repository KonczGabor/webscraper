package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Product struct {
	Name, Price, Image, URL string
}

func main() {
	// initialize the Chrome instance
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	var products []Product

	// create a channel to receive products
	productChan := make(chan Product)
	done := make(chan bool)

	// start a goroutine to collect products
	go func() {
		for product := range productChan {
			products = append(products, product)
		}
		done <- true
	}()

	// navigate and scrape
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.scrapingcourse.com/infinite-scrolling"),
		scrapeProducts(productChan),
	)
	if err != nil {
		log.Fatal(err)
	}

	close(productChan)
	<-done

	// print results
	fmt.Printf("Scraped %d products\n", len(products))
	for _, p := range products {
		fmt.Printf("Name: %s\nPrice: %s\nImage: %s\nURL: %s\n\n",
			p.Name, p.Price, p.Image, p.URL)
	}
}

func scrapeProducts(productChan chan<- Product) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var previousHeight int
		for {
			// get all product nodes
			var nodes []*cdp.Node
			if err := chromedp.Nodes(".product-item", &nodes).Do(ctx); err != nil {
				return err
			}

			// extract data from each product
			for _, node := range nodes {
				var product Product

				// using chromedp's node selection to extract data
				if err := chromedp.Run(ctx,
					chromedp.Text(".product-name", &product.Name, chromedp.ByQuery, chromedp.FromNode(node)),
					chromedp.Text(".product-price", &product.Price, chromedp.ByQuery, chromedp.FromNode(node)),
					chromedp.AttributeValue("img", "src", &product.Image, nil, chromedp.ByQuery, chromedp.FromNode(node)),
					chromedp.AttributeValue("a", "href", &product.URL, nil, chromedp.ByQuery, chromedp.FromNode(node)),
				); err != nil {
					continue
				}

				// clean price text
				product.Price = strings.TrimSpace(product.Price)

				// send product to channel if not empty
				if product.Name != "" {
					productChan <- product
				}
			}

			// scroll to bottom
			var height int
			if err := chromedp.Evaluate(`document.documentElement.scrollHeight`, &height).Do(ctx); err != nil {
				return err
			}

			// break if we've reached the bottom (no height change after scroll)
			if height == previousHeight {
				break
			}
			previousHeight = height

			// scroll and wait for content to load
			fmt.Println("New page loading...")
			if err := chromedp.Run(ctx,
				chromedp.Evaluate(`window.scrollTo(0, document.documentElement.scrollHeight)`, nil),
				chromedp.Sleep(1*time.Second), // Wait for new content to load
			); err != nil {
				return err
			}

		}
		return nil
	}
}
