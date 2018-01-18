package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type row struct {
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

var (
	ts = httptest.NewServer(http.HandlerFunc(SearchServer))
)

type root struct {
	Rows []row `xml:"row"`
}

type TestCase struct {
	Request SearchRequest
	Result  Result
}

type Result struct {
	Response *SearchResponse
	Err      error
}

type ByID []User

type ByAge []User

type ByName []User

func (s ByID) Len() int {
	return len(s)
}
func (s ByID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByID) Less(i, j int) bool {
	return s[i].Id < s[j].Id
}

func (s ByAge) Len() int {
	return len(s)
}
func (s ByAge) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByAge) Less(i, j int) bool {
	return s[i].Age < s[j].Age
}

func (s ByName) Len() int {
	return len(s)
}
func (s ByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByName) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func reverse(users []User) {
	for i, j := 0, len(users)-1; i < j; i, j = i+1, j-1 {
		users[i], users[j] = users[j], users[i]
	}
}

const (
	token = "secret"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		fmt.Errorf("Can't open file dataset.xml")
	}
	v := new(root)
	xml.Unmarshal(data, &v)
	//limit := r.FormValue("limit")
	//offset := r.FormValue("offset")
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		fmt.Errorf("error converting")
	}
	rows := v.Rows
	var users []User
	for _, x := range rows {
		name := x.FirstName + " " + x.LastName
		if strings.Contains(name, query) || strings.Contains(x.About, query) || len(query) == 0 {
			user := User{x.ID, name, x.Age, x.About, x.Gender}
			users = append(users, user)
		}
	}
	if orderBy != 0 {
		switch orderField {
		case "ID":
			sort.Sort(ByID(users))
		case "Age":
			sort.Sort(ByAge(users))
		case "Name", "":
			sort.Sort(ByName(users))
		default:
			w.WriteHeader(http.StatusBadRequest)
			js, err := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
			if err != nil {
				fmt.Errorf("can't marshal error")
			}
			w.Write(js)
			return
		}
		if orderBy == -1 {
			reverse(users)
		}
		if orderBy != -1 && orderBy != 1 {
			w.WriteHeader(http.StatusBadRequest)
			js, err := json.Marshal(SearchErrorResponse{"can't choose sort style"})
			if err != nil {
				fmt.Errorf("can't marshal error")
			}
			w.Write(js)
			return
		}
	}
	accessToken := r.Header.Get("AccessToken")
	if accessToken != token {
		w.WriteHeader(http.StatusUnauthorized)
		js, err := json.Marshal(SearchErrorResponse{"Bad AccessToken"})
		if err != nil {
			fmt.Errorf("can't marshal error")
		}
		w.Write(js)
		return
	}
	if r.URL.Path == "/help" {
		w.WriteHeader(http.StatusBadRequest)
		js, err := json.Marshal("help doesn't exist")
		if err != nil {
			fmt.Errorf("can't marshal error")
		}
		w.Write(js)
		return
	}
	if r.URL.Path == "/redirect" {
		w.WriteHeader(http.StatusFound)
		//http.Redirect(w, r, r.URL.Path+"/someurl", http.StatusFound)
		return
	}
	if r.URL.Path == "/timeout" {
		w.WriteHeader(http.StatusFound)
		time.Sleep(2 * time.Second)
	}
	if r.URL.Path == "/about" {
		w.WriteHeader(http.StatusOK)
		js, err := json.Marshal("All rights reserved")
		if err != nil {
			fmt.Errorf("can't marshal error")
		}
		w.Write(js)
		return
	}
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusInternalServerError)
		js, err := json.Marshal(SearchErrorResponse{"SearchServer fatal error"})
		if err != nil {
			fmt.Errorf("can't marshal error")
		}
		w.Write(js)
		return
	}
	w.WriteHeader(http.StatusOK)
	js, err := json.Marshal(users)
	if err != nil {
		fmt.Errorf("can't marshal answer")
	}
	w.Write(js)
}

func TestFindUsersNextPage(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      3,
			Offset:     0,
			Query:      "J",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			&SearchResponse{
				Users: []User{
					User{
						Id:     25,
						Name:   "Katheryn Jacobs",
						Age:    32,
						About:  "Magna excepteur anim amet id consequat tempor dolor sunt id enim ipsum ea est ex. In do ea sint qui in minim mollit anim est et minim dolore velit laborum. Officia commodo duis ut proident laboris fugiat commodo do ex duis consequat exercitation. Ad et excepteur ex ea exercitation id fugiat exercitation amet proident adipisicing laboris id deserunt. Commodo proident laborum elit ex aliqua labore culpa ullamco occaecat voluptate voluptate laboris deserunt magna.\n",
						Gender: "female",
					},
					User{
						Id:     21,
						Name:   "Johns Whitney",
						Age:    26,
						About:  "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n",
						Gender: "male",
					},
					User{
						Id:     6,
						Name:   "Jennings Mays",
						Age:    39,
						About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			nil,
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(testCase.Request)
	if err != nil || !reflect.DeepEqual(testCase.Result.Response, result) {
		t.Errorf("wrong result, \n\nexpected result: %#v\n error: %#v,\n\n got result: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersLimitLess(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      -5,
			Offset:     0,
			Query:      "Jennings",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("limit must be > 0"),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersOffset(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      5,
			Offset:     -1,
			Query:      "Jennings",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("offset must be > 0"),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersLimitMore(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "o",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: &SearchResponse{
				Users:    []User{},
				NextPage: false,
			},
			Err: nil,
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL,
	}
	result, err := s.FindUsers(testCase.Request)
	if result == nil || err != nil || result.NextPage != false {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersBadRequest(t *testing.T) {
	testCases := []TestCase{
		TestCase{
			Request: SearchRequest{
				Limit:      5,
				Offset:     5,
				Query:      "Jennings",
				OrderField: "Bad",
				OrderBy:    OrderByAsc,
			},
			Result: Result{
				Response: nil,
				Err:      errors.New("OrderFeld Bad invalid"),
			},
		},
		TestCase{
			Request: SearchRequest{
				Limit:      5,
				Offset:     5,
				Query:      "Jennings",
				OrderField: "",
				OrderBy:    7,
			},
			Result: Result{
				Response: nil,
				Err:      errors.New("unknown bad request error: can't choose sort style"),
			},
		},
	}

	for _, x := range testCases {
		s := &SearchClient{
			AccessToken: token,
			URL:         ts.URL,
		}
		result, err := s.FindUsers(x.Request)
		if result != nil || err.Error() != x.Result.Err.Error() {
			t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", x.Result.Response, x.Result.Err, result, err)
		}
	}
}

func TestFindUsersUnauthorized(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("Bad AccessToken"),
		},
	}

	s := &SearchClient{
		AccessToken: token + "salt",
		URL:         ts.URL,
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersInternalServerError(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("SearchServer fatal error"),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL + "/xss",
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersBadRequestHelp(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("cant unpack error json: json: cannot unmarshal string into Go value of type main.SearchErrorResponse"),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL + "/help",
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersBadRequestAbout(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("cant unpack result json: json: cannot unmarshal string into Go value of type []main.User"),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL + "/about",
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersRedirect(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("unknown error Get "),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL + "/redirect",
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || !strings.Contains(err.Error(), testCase.Result.Err.Error()) {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}

func TestFindUsersTimeout(t *testing.T) {
	testCase := TestCase{
		Request: SearchRequest{
			Limit:      50,
			Offset:     1,
			Query:      "",
			OrderField: "",
			OrderBy:    OrderByAsc,
		},
		Result: Result{
			Response: nil,
			Err:      errors.New("timeout for limit=26&offset=1&order_by=-1&order_field=&query="),
		},
	}

	s := &SearchClient{
		AccessToken: token,
		URL:         ts.URL + "/timeout",
	}
	result, err := s.FindUsers(testCase.Request)
	if result != nil || err.Error() != testCase.Result.Err.Error() {
		t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", testCase.Result.Response, testCase.Result.Err, result, err)
	}
}
