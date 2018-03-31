package main

import (
	"errors"
	"html/template"
	"net/http"
)

var templates map[string]*template.Template

var ErrTemplateDoesNotExist = errors.New("The template does not exist.")

// renderTemplate is a wrapper around template.ExecuteTemplate.
func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		return ErrTemplateDoesNotExist
	}

	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.ExecuteTemplate(w, "base", data)

	return nil
}

func init() {
	// Init templates
	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	templates["index.html"] = template.Must(template.ParseFiles("templates/base.html", "templates/index.html", "templates/navigation.html"))
	templates["login.html"] = template.Must(template.ParseFiles("templates/base.html", "templates/login.html", "templates/navigation.html"))
	templates["information.html"] = template.Must(template.ParseFiles("templates/base.html", "templates/information.html", "templates/navigation.html"))
	templates["snapshot.html"] = template.Must(template.ParseFiles("templates/base.html", "templates/snapshot.html", "templates/navigation.html"))
	templates["video.html"] = template.Must(template.ParseFiles("templates/base.html", "templates/video.html", "templates/navigation.html"))
}
