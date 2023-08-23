package main

import (
	"net/http"
	"url_shortner/internal/data"
)

type application struct {
	URL *data.URLModel
}

func main() {

	app := application{
		URL: &data.URLModel{
			DB: make([]*data.URL, 0),
		},
	}

	http.ListenAndServe(":8080", app.routes())
}
