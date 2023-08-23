package data

import (
	"fmt"
	"url_shortner/internal/utils"
)

type URL struct {
	Long     string
	Short    string
	Accessed int64
}

func NewURL(longURL string) *URL {
	shortURL := utils.Shorten(longURL)
	return &URL{
		Long:     longURL,
		Short:    shortURL,
		Accessed: 0,
	}
}

type URLModel struct {
	DB []*URL
}

func (model *URLModel) Insert(url *URL) error {
	model.DB = append(model.DB, url)
	return nil
}

func (model *URLModel) Get(shortURL string) *URL {
	for i, url := range model.DB {
		if url.Short == shortURL {
			url.Accessed++
			fmt.Println("Found", url)
			return model.DB[i]
		} else {
			fmt.Println("urls are", url.Short, shortURL)
		}
	}
	return nil
}
