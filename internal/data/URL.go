package data

import "fmt"

type URL struct {
	Long     string
	Short    string
	Accessed int64
}

func NewURL(longURL string) *URL {
	shortURL := shorten(longURL)
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
	fmt.Println("models", model.DB)
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
