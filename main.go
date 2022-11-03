package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gocolly/colly"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Item struct {
	Name              string
	OldPrice          string
	BlackFridayPrice  string
	Difference        string
	DifferencePercent string
	Url               string
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

		// Removes the accents from the string
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		name, _, _ := transform.String(t, e.ChildAttr("img", "alt"))

		rePrice := regexp.MustCompile(`R\$[0-9]+,[0-9]+`)
		pricesMatches := rePrice.FindAllString(e.ChildText("span.price-tag-amount"), -1)
		replacer := strings.NewReplacer("R$", "", ",", ".")

		for i, price := range pricesMatches {
			pricesMatches[i] = replacer.Replace(price)
		}

		// concatenate 'R$' to the old price and black friday price
		oldPrice, blackFridayPrice := pricesMatches[0], pricesMatches[1]
		difference, differencePercent := getDifference(oldPrice, blackFridayPrice)

		items = append(items, Item{
			Name:              name,
			OldPrice:          fmt.Sprintf("R$%s", oldPrice),
			BlackFridayPrice:  fmt.Sprintf("R$%s", blackFridayPrice),
			Difference:        difference,
			DifferencePercent: differencePercent,
			Url:               e.ChildAttr("a", "href"),
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

func getDifference(oldPrice, blackFridayPrice string) (string, string) {

	oldPriceFloat, _ := strconv.ParseFloat(oldPrice, 64)
	blackFridayPriceFloat, _ := strconv.ParseFloat(blackFridayPrice, 64)

	difference := oldPriceFloat - blackFridayPriceFloat
	differencePercent := (difference / oldPriceFloat) * 100

	// concatenate 'R$' to the difference
	differenceStr := fmt.Sprintf("R$%.2f", difference)
	// concatenate '%' to the difference percent
	differencePercentStr := fmt.Sprintf("%.2f%%", differencePercent)

	return differenceStr, differencePercentStr
}

func writeCSV(items []Item) {

	file, err := os.Create("dados.csv")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	werr := writer.Write([]string{"Produto", "Preco Antigo", "Preco Black Friday", "Diferenca de Preco", "Diferenca Percentual", "Pagina do Produto"})
	if werr != nil {
		log.Fatal("Cannot write to file", err)
	}

	for _, item := range items {
		err := writer.Write([]string{item.Name, item.OldPrice, item.BlackFridayPrice, item.Difference, item.DifferencePercent, item.Url})
		if err != nil {
			log.Fatal("Cannot write to file", err)
		}
	}
}
