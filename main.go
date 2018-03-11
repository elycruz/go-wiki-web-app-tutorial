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

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "index.html", "error.html"))
var isValidPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z\\d]+)$")

type Any interface {}

type Page struct {
	Title string
	Body []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func getTitle (w http.ResponseWriter, r *http.Request) (string, error) {
	m := isValidPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // title is second sub expression
}

func loadPage (title string) (*Page, error){
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename);
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func executeServerInternalError (res http.ResponseWriter, err error) {
	http.Error(res, err.Error(), http.StatusInternalServerError)
}

func renderPageTemplate (res http.ResponseWriter, action string, p *Page) {
	err := templates.ExecuteTemplate(res, action + ".html", p)
	if err != nil {
		executeServerInternalError(res, err)
		return
	}
}

func viewHandler (res http.ResponseWriter, req *http.Request) {
	title := req.URL.Path[len("/view/"):]
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(res, req, "/edit/" + title, http.StatusNotFound);
		return
	}
	t, _ := template.ParseFiles("view.html")
	err = t.Execute(res, *page)
	if err != nil {
		executeServerInternalError(res, err)
	}
}

func indexHandler (res http.ResponseWriter, req *http.Request) {
	renderPageTemplate(
		res,
		"index",
		&Page{"Index", []byte("Hello World")})
}

func editHandler (res http.ResponseWriter, req *http.Request) {
	title := req.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title, Body: []byte("Page doesn't exist.  Opened edit screen for editing.")}
	}
	t, _ := template.ParseFiles("edit.html")
	t.Execute(res, p)
}

func saveHandler (w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/save/"):]
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		executeServerInternalError(w, err)
		return;
	}
	http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

func ensureTestPage () {
	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
	p1.save()
}

func main () {
	ensureTestPage()

	// Set handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	// Listen on 8080
	fmt.Println("Listening on port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
