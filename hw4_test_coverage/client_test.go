package main

import (
	"encoding/xml"
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
	// распарсить запросу
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy, _ := strconv.Atoi("order_by")

	// распарсить
	dataset, _ := ioutil.ReadFile("dataset.xml")

	root := new(Root)
	xml.Unmarshal(dataset, &root)

	// отфильтровать
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

	// отсортировать
	if orderField == "Id" {
		if orderBy == -1 {
			sort.Sort(ByIdAsc(rows))
		} else if orderBy == 1 {
			sort.Sort(ByIdDesc(rows))
		}
	} else if orderField == "Age" {
		if orderBy == -1 {
			sort.Sort(ByAgeAsc(rows))
		} else if orderBy == 1 {
			sort.Sort(ByAgeDesc(rows))
		}
	} else if orderField == "Name" || orderField == "" {
		if orderBy == -1 {
			sort.Sort(ByNameAsc(rows))
		} else if orderBy == 1 {
			sort.Sort(ByNameDesc(rows))
		}
	} else {
		// вернуть ошибку
	}

	// отрезать
	result := rows[offset : offset+limit]

	// вернуть результат
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
