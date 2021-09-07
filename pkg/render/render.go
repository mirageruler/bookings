package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/mirageruler/bookings/pkg/config"
	"github.com/mirageruler/bookings/pkg/models"
)

var functions = template.FuncMap{}

var app *config.AppConfig

// NewTemplates sets the config for the template package
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData) *models.TemplateData {

	return td
}

// RenderTemplate renders templates using html/template
func RenderTemplate(w http.ResponseWriter, html string, td *models.TemplateData) {
	var tc map[string]*template.Template
	// If we're in development mode, don't use the template cache, just rebuild one everytime a page is loaded,
	// otherwise, let's just use the template cache for production mode.
	if app.UseCache {
		// Get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[html]
	if !ok {
		log.Fatal("Could not get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writting template to browser", err)
	}
}

// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	// This just creates an empty map
	myCache := map[string]*template.Template{}

	// This gets a list of all files ending with page.html, and stores it in a slice of strings called 'pages'
	pages, err := filepath.Glob("./templates/*.page.html")
	if err != nil {
		return myCache, err
	}

	// Now we loop through the slice of strings, which has two entries: "home.page.html" and "about.page.html"
	for _, page := range pages {
		name := filepath.Base(page)
		// templates set
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		// Here, we are checking to see if there are any files at all that end with layout.html.
		// There is only one, but if there were more than one, we will get them all and store them
		// in a slice of strings called 'matches'
		matches, err := filepath.Glob("./templates/*.layout.html")
		if err != nil {
			return myCache, err
		}

		// If the length of matches is > 0, we go through the slice and parse all of the layouts
		// available to us. We might not use any of them in this iteration through the loop, but if
		// the current template we are working on (home.page.html the first time through) does use a layout,
		// we need to have it available to us before we add it to our template set.
		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.html")
			if err != nil {
				return myCache, err
			}
		}

		// The first time through, name is still home.page.html, we never add anything with *.layout.html to the
		// template set; we just use the layout to create a page which depends on it.
		// Now, we add the templates, complete any associated layouts, to our template set.
		myCache[name] = ts
	}
	return myCache, nil
}
