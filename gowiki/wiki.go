package main

import (
    "fmt"
    "os"
    "log"
    "net/http"
    "html/template"
    "regexp"
    "errors"
)

type Page struct {
    Title string
    Body []byte
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
func (p *Page) save() error {
    filename := p.Title + ".txt"
    return os.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    log.Printf(fmt.Sprintf("Loading page: %s", title))
    filename := title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, templateName string, p *Page) {
    err := templates.ExecuteTemplate(w, fmt.Sprintf("%s.html", templateName), p) 
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func getPageTitle(r *http.Request, templateName string) (string) {
    toExclude := len(fmt.Sprintf("/%s/", templateName))
    return r.URL.Path[toExclude:]
}
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("invalid page title")
    }
    return m[2], nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    log.Printf("viewHandler")
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    log.Printf("editHandler")
    p, err := loadPage(title) 
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    log.Printf("saveHandler")
    body := r.FormValue("body") 
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, fmt.Sprintf("/view/%s", title), http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return 
        }
        fn(w, r, m[2])
    }
}
func main() {
    //p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page!")}
    //p1.save()
    //p2, _ := loadPage("TestPage")
    //fmt.Println(string(p2.Body))
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))

    log.Fatal(http.ListenAndServe(":8080", nil))
}
