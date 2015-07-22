package main

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Scrape(word string) *Entry {
	r1, err := http.Get("http://lema.rae.es/drae/srv/search?val=" + word)

	if err != nil {
		panic(err)
	}

	r2 := Solve(r1)
	defer r2.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(r2)
	if err != nil {
		panic(err)
	}

	nodes := doc.Find("body").Children().Filter("div")
	etymology := strings.TrimSpace(nodes.First().Find("span.a").Text())
	defs := []*Definition{}
	vars := []*Variation{}

	nodes.Each(func(i int, s *goquery.Selection) {
		delimiter := s.Children().Filter("p:not([class])").First()
		delimiter.NextAll().EachWithBreak(func(i int, s *goquery.Selection) bool {
			//Skip empty P tags.
			if s.Children().Length() == 0 {
				return true
			}
			//Break when the first variation is reached
			if s.HasClass("p") {
				return false
			}
			defs = append(defs, ScrapeDefinition(s))
			return true
		})

		delimiters := s.Find(".p span.k").Parent()
		delimiters.Each(func(i int, v *goquery.Selection) {
			vars = append(vars, &Variation{Variation: strings.TrimSpace(v.Text())})
			v.NextAll().EachWithBreak(func(j int, s *goquery.Selection) bool {
				//Done with this variation.
				if s.HasClass("p") {
					return false
				}
				vars[i].Definitions = append(vars[i].Definitions, ScrapeDefinition(s))
				return true
			})
		})
	})

	entry := &Entry{
		Word:        word,
		Etymology:   etymology,
		Definitions: defs,
		Variations:  vars,
	}

	return entry
}

func ScrapeDefinition(s *goquery.Selection) *Definition {
	category, _ := s.Find("span[title]").First().Attr("title")

	def := &Definition{
		Category:   category,
		Definition: strings.TrimSpace(s.Find("span.b").Clone().Children().Not("a").Remove().End().End().Text()),
		Origin:     ScrapeOrigins(s),
		Notes:      ScrapeNotes(s),
		Examples:   ScrapeExamples(s),
	}

	return def
}

func ScrapeOrigins(s *goquery.Selection) []string {
	origins := []string{}
	s.Find("span.d i span.d[title]").Each(func(i int, s *goquery.Selection) {
		origin, _ := s.Attr("title")
		origins = append(origins, origin)
	})
	return origins
}

func ScrapeNotes(s *goquery.Selection) []string {
	notes := []string{}
	s.Clone().Find("span[title]").First().Remove().End().End().Find("span.d i span.d[title]").Remove().End().Find("span.d[title]").Each(func(i int, s *goquery.Selection) {
		note, _ := s.Attr("title")
		notes = append(notes, note)
	})
	return notes
}

func ScrapeExamples(s *goquery.Selection) []string {
	examples := []string{}
	s.Find("span.h i").Each(func(i int, s *goquery.Selection) {
		examples = append(examples, s.Text())
	})
	return examples
}
