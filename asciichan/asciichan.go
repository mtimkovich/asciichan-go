package asciichan

import (
    "appengine"
    "appengine/datastore"

    "net/http"
    "html/template"

    "unicode"
    "time"
)

type Art struct {
    Title string
    Art []byte
    Created time.Time
}

type PrettyArt struct {
    Title string
    Art string
    Created time.Time
}

type Data struct {
    FormTitle string
    FormArt string
    Error string

    Arts []Art

    PrettyArts []PrettyArt
}

func isBlank(input string) bool {
    if len(input) == 0 {
        return true
    } 

    for _, c := range input {
        if !unicode.IsSpace(c) {
            return false
        }
    }

    return true
}

func renderFront(c appengine.Context, w http.ResponseWriter, data *Data) {
    if data == nil {
        data = new(Data)
    }

    q := datastore.NewQuery("Art").Order("-Created").Limit(10)

    data.Arts = make([]Art, 0, 10)
    data.PrettyArts = make([]PrettyArt, 0, 10)

    if _, err := q.GetAll(c, &data.Arts); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    for _, v := range data.Arts {
        p := PrettyArt{
            Title: v.Title,
            Art: string(v.Art),
            Created: v.Created,
        }

        data.PrettyArts = append(data.PrettyArts, p)
    }

    templates := template.Must(template.ParseFiles("templates/index.html"))
    templates.ExecuteTemplate(w, "index.html", *data)
}

func Root(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    if r.Method == "GET" {
        renderFront(c, w, nil)

    } else if r.Method == "POST" {
        data := new(Data)

        data.FormTitle = r.FormValue("title")
        data.FormArt = r.FormValue("art")

        if isBlank(data.FormTitle) || isBlank(data.FormArt) {
            data.Error = "title and art cannot be blank"
            
            renderFront(c, w, data)
        } else {
            a := Art{
                Title: data.FormTitle,
                Art: []byte(data.FormArt),
                Created: time.Now(),
            }

            _, err := datastore.Put(c, datastore.NewIncompleteKey(c, "Art", nil), &a)

            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            c.Infof("INSERT INTO Art (Title, Art, Created) values ('%v', %v', %v)", a.Title, a.Art, a.Created)

            http.Redirect(w, r, "/", http.StatusFound)
        }

    }
}

func init() {
    http.HandleFunc("/", Root)
}
