/**
 * Go-Wiki Web Application Tutorial
 * @see https://golang.org/doc/articles/wiki/
 */
package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"errors"
	"log"
)

var templates = template.Must(template.ParseFiles(
		"views/edit.html",
		"views/view.html",
		"views/index.html",
		"views/error.html"))

var createdPagesPath = "./created-pages"

var isValidPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z\\d]+)$")

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := createdPagesPath + "/" + p.Title + ".html"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func pathParts (path string) ([]string) {
	return isValidPath.FindStringSubmatch(path)
}

func possiblePathParts (path string) ([]string, error) {
	m := pathParts(path)
	if m == nil {
		return nil, errors.New("Invalid application path: " + path)
	}
	return m, nil
}

func getPageIdOrError (w http.ResponseWriter, r *http.Request) (string, error) {
	parts, err := possiblePathParts(r.URL.Path)
	if err != nil {
		return "", err
	}
	return parts[2], nil
}

func serverInternalError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func loadPageOrError (title string) (*Page, error){
	filename := createdPagesPath + "/" + title + ".html"
	body, err := ioutil.ReadFile(filename);
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderPageTemplateOrStatus500 (w http.ResponseWriter, action string, p *Page) {
	err := templates.ExecuteTemplate(w, action + ".html", p)
	if err != nil {
		serverInternalError(w, err)
	}
}

func makeHandler (fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		fmt.Printf("[%s] - %s\n", r.Method, path)
		if path == "/index" || path == "/" {
			fn(w, r, "index")
			return
		}

		id, err := getPageIdOrError(w, r)
		if err != nil {
			http.NotFound(w, r)
			return;
		}

		fn(w, r, id)
	}
}

func viewHandler (w http.ResponseWriter, req *http.Request,pageId string) {
	page, err := loadPageOrError(pageId)
	if err != nil {
		http.Redirect(w, req, "/edit/" + pageId, http.StatusNotFound);
		return
	}
	renderPageTemplateOrStatus500(w, "view", page)
}

func indexHandler (w http.ResponseWriter, req *http.Request, pageId string) {
	renderPageTemplateOrStatus500(
		w,
		"index",
		&Page{"Index", []byte("Hello World")})
}

func editHandler (w http.ResponseWriter, req *http.Request, pageId string) {
	p, err := loadPageOrError(pageId)
	if err != nil {
		p = &Page{Title: pageId, Body: []byte("Page doesn't exist.  Opened edit screen for editing.")}
	}
	renderPageTemplateOrStatus500(w, "edit", p)
}

func saveHandler (w http.ResponseWriter, r *http.Request, pageId string) {
	body := r.FormValue("body")
	p := &Page{Title: pageId, Body: []byte(body)}
	err := p.save()
	if err != nil {
		serverInternalError(w, err)
		return;
	}
	http.Redirect(w, r, "/view/" + pageId, http.StatusFound)
}

func ensureTestPage () {
	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
	p1.save()
}

func main () {
	ensureTestPage()

	// Set handlers
	http.HandleFunc("/", 		makeHandler(indexHandler))
	http.HandleFunc("/view/", 	makeHandler(viewHandler))
	http.HandleFunc("/edit/", 	makeHandler(editHandler))
	http.HandleFunc("/save/", 	makeHandler(saveHandler))

	// Listen on 8080
	fmt.Println("Listening on port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
