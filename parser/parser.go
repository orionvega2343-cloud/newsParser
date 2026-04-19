package parser

import (
	"database/sql"
	"fmt"
	"newsParser/models"
	"newsParser/storage"

	"github.com/gocolly/colly"
)

func Scrape(URL string, db *sql.DB) []models.Result {
	c := colly.NewCollector(colly.Async(true))                      //Одновременный парсинг
	var results []models.Result                                     //Пустой слайс
	c.OnHTML("a.cell-list__item-link", func(e *colly.HTMLElement) { //Поиску по тегу
		r := models.Result{Header: e.Text, Link: e.Attr("href")} //Получение текста
		res, err := storage.Insert(db, r)
		if err != nil {
			fmt.Println(err)
		}
		results = append(results, res) //Добавляем в слайс

	})
	c.Visit(URL) //Заходим на сайт
	c.Wait()     //Что то ждем
	return results
}
