package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/gocolly/colly"
)

type Item struct {
	Name             string
	OldPrice         string
	BlackFridayPrice string
	Url              string
}

func main() {

	items := scraper("https://lista.mercadolivre.com.br/_Deal_plantao-black-friday-mais-vendidos")
	writeCSV(items)

}

func scraper(url string) []Item {

	items := []Item{}

	c := colly.NewCollector()
	c.OnHTML("ol.ui-search-layout", func(e *colly.HTMLElement) {
		if e == nil {
			log.Fatal("The page content is empty")
		}

		rePrice := regexp.MustCompile(`R\$\d+[\.\,]?\d*`)
		prices := rePrice.FindAllString(e.ChildText("span.price-tag-amount"), -1)
		items = append(items, Item{
			Name:             e.ChildAttr("img", "alt"),
			OldPrice:         prices[0],
			BlackFridayPrice: prices[1],
			Url:              e.ChildAttr("a", "href"),
		})

	})

	c.OnHTML("li.andes-pagination__page-count", func(e *colly.HTMLElement) {
		if e == nil {
			log.Fatal("The page content is empty")
		}

		// reNumPages := regexp.MustCompile(`(?m)\s(\d+)`)
		// numPages := reNumPages.FindString(e.Text)[1:]
	})

	c.OnHTML("li.andes-pagination__button.andes-pagination__button--next.shops__pagination-button", func(e *colly.HTMLElement) {
		nextPage := e.ChildAttr("a", "href")
		if nextPage != "" {
			c.Visit(nextPage)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(url)

	return items
}

func writeCSV(items []Item) {
	file, err := os.Create("items.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, item := range items {
		err := writer.Write([]string{item.Name, item.OldPrice, item.BlackFridayPrice, item.Url})
		if err != nil {
			log.Fatal("Cannot write to file", err)
		}
	}
}
