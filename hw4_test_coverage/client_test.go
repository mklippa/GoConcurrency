package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type Root struct {
	XMLName xml.Name `xml:"root"`
	Text    string   `xml:",chardata"`
	Rows    []Row    `xml:"row"`
}

type Row struct {
	Text          string `xml:",chardata"`
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type ByIdAsc []Row

func (a ByIdAsc) Len() int           { return len(a) }
func (a ByIdAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIdAsc) Less(i, j int) bool { return a[i].ID < a[j].ID }

type ByAgeAsc []Row

func (a ByAgeAsc) Len() int           { return len(a) }
func (a ByAgeAsc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAgeAsc) Less(i, j int) bool { return a[i].Age < a[j].Age }

type ByNameAsc []Row

func (a ByNameAsc) Len() int      { return len(a) }
func (a ByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNameAsc) Less(i, j int) bool {
	return a[i].FirstName+a[i].LastName < a[j].FirstName+a[j].LastName
}

type ByIdDesc []Row

func (a ByIdDesc) Len() int           { return len(a) }
func (a ByIdDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIdDesc) Less(i, j int) bool { return a[i].ID > a[j].ID }

type ByAgeDesc []Row

func (a ByAgeDesc) Len() int           { return len(a) }
func (a ByAgeDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAgeDesc) Less(i, j int) bool { return a[i].Age > a[j].Age }

type ByNameDesc []Row

func (a ByNameDesc) Len() int      { return len(a) }
func (a ByNameDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNameDesc) Less(i, j int) bool {
	return a[i].FirstName+a[i].LastName > a[j].FirstName+a[j].LastName
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// авторизация
	if r.Header.Get("AccessToken") != "test" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// получение параметров
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy, _ := strconv.Atoi("order_by")

	// парсинг данных
	dataset, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	root := new(Root)
	err = xml.Unmarshal(dataset, &root)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// фильтрация
	rows := make([]Row, 0)
	if query != "" {
		for _, row := range root.Rows {
			if strings.Contains(row.About, query) || strings.Contains(row.FirstName+row.LastName, query) {
				rows = append(rows, row)
			}
		}
	} else {
		rows = root.Rows
	}

	// сортировка
	if orderField == "Id" {
		if orderBy == OrderByAsc {
			sort.Sort(ByIdAsc(rows))
		} else if orderBy == OrderByDesc {
			sort.Sort(ByIdDesc(rows))
		} else if orderBy == OrderByAsIs {
		} else {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"Error": "ErrorBadOrderBy"}`)
			return
		}
	} else if orderField == "Age" {
		if orderBy == OrderByAsc {
			sort.Sort(ByAgeAsc(rows))
		} else if orderBy == OrderByDesc {
			sort.Sort(ByAgeDesc(rows))
		} else if orderBy == OrderByAsIs {
		} else {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"Error": "ErrorBadOrderBy"}`)
			return
		}
	} else if orderField == "Name" || orderField == "" {
		if orderBy == OrderByAsc {
			sort.Sort(ByNameAsc(rows))
		} else if orderBy == OrderByDesc {
			sort.Sort(ByNameDesc(rows))
		} else if orderBy == OrderByAsIs {
		} else {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"Error": "ErrorBadOrderBy"}`)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "ErrorBadOrderField"}`)
		return
	}

	// пагинация
	if offset >= len(rows) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"Error": "OffsetOutOfRange"}`)
		return
	}
	last := offset + limit
	if last > len(rows) {
		last = len(rows)
	}
	rows = rows[offset:last]

	// ответ
	result := make([]User, 0)
	for _, row := range rows {
		user := User{
			Id:     row.ID,
			Name:   row.FirstName + row.LastName,
			Gender: row.Gender,
			Age:    row.Age,
			About:  row.About,
		}
		result = append(result, user)
	}

	s, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, string(s))
}

func TestSearchServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	c := &SearchClient{
		AccessToken: "",     //! что установить?
		URL:         ts.URL, //! возможно, что-то не то установил
	}

	c.FindUsers(SearchRequest{})

	ts.Close()
}

/*
Тест-кейсы:
	- limit < 0 -> error
	- limit > 25 -> return 25 items
	- offset < 0 -> error
	- (?) limit is not int -> panic
	- (?) offset is not int -> panic
	- (?) order_by is not int -> panic
	- token is not eq "test" -> auth error
	- request timeout error (alter. token(?))
	- unknown error
	- one of internal error
	- can't unpack dad request response
	- bad order field
	- unknown bad request (bad order by ?)
	- can't unpack result json
	- resp len == limit -> next page == true
	- (?) resp len == limit -> resp OK
	- resp len <> limit -> next page == false
	- postitive case
*/
