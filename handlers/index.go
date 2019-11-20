package handlers

import (
	"html/template"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("index.html").ParseFiles("frontend/index.html")
	if err != nil {
		log.WithError(err).Errorf("Can't parse index.html")
		http.Error(w, "Template error", http.StatusServiceUnavailable)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		log.WithError(err).Errorf("Can't execute index.html")
		http.Error(w, "Template error", http.StatusServiceUnavailable)
		return
	}
}
