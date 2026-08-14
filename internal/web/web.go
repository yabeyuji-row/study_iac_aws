package web

import (
	"html/template"
	"net/http"
)

const (
	staticDir     = "internal/web/static"
	indexTemplate = "internal/web/templates/index.html"
)

type IndexData struct {
	Lang       string
	Title      string
	Heading    string
	StylesPath string
	ScriptPath string
	Statuses   []StatusOption
}

type StatusOption struct {
	Value string
	Label string
}

func Handler() (handler http.Handler) {
	staticHandler := http.FileServer(http.Dir(staticDir))
	parsedTemplate, err := template.ParseFiles(indexTemplate)
	if err != nil {
		return http.NotFoundHandler()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			staticHandler.ServeHTTP(response, request)
			return
		}

		response.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := parsedTemplate.Execute(response, defaultIndexData()); err != nil {
			http.Error(response, "failed to render template", http.StatusInternalServerError)
		}
	})

	return mux
}

func defaultIndexData() (data IndexData) {
	return IndexData{
		Lang:       "ja",
		Title:      "TODO API",
		Heading:    "TODO",
		StylesPath: "/styles.css",
		ScriptPath: "/app.js",
		Statuses: []StatusOption{
			{Value: "pending", Label: "Pending"},
			{Value: "in_progress", Label: "In progress"},
			{Value: "completed", Label: "Completed"},
		},
	}
}
