package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/thedevsaddam/gojsonq"
)

type adNetwork struct {
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Value       int    `json:"value"`
	Platform    string `json:"platform"`
	OsVersion   string `json:"osversion"`
	AppName     string `json:"appname"`
	AppVersion  string `json:"appversion"`
	CountryCode string `json:"countrycode"`
	AdType      string `json:"adtype"`
}

var adNetworks []adNetwork

var filePath = "output.txt"

func returnAllAdNetworks(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "	")
	enc.Encode(adNetworks)
}

func returnAdType(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	adtype := vars["adtype"]

	var list []adNetwork

	for _, adNetwork := range adNetworks {
		if strings.ToLower(adNetwork.AdType) == strings.ToLower(adtype) {
			list = append(list, adNetwork)
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "	")
	enc.Encode(list)
}

func queryAdNetworks(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	platform, platformErr := params["platform"]
	osversion, osversionErr := params["osversion"]

	jq := gojsonq.New().File(filePath)
	res := jq.WhereNotNil("id")

	i := 0
	for param, valueParam := range params {
		if strings.ToLower(param) == "id" {
			value, err := strconv.ParseInt(valueParam[0], 10, 64)
			if err != nil {
				fmt.Println("Query ERROR")
				return
			}
			res = jq.WhereEqual(strings.ToLower(param), value)
		} else {
			value := valueParam[0]
			res = jq.WhereEqual(strings.ToLower(param), strings.ToLower(value))
		}

		if platformErr == true && osversionErr == true && i == 0 {
			i++
			if strings.ToLower(platform[0]) == "android" && strings.ToLower(osversion[0]) == "9" {
				res = jq.WhereNotEqual("description", "admob")
			}
		}
	}

	resGet := res.SortBy("value", "desc").Get()
	file, _ := json.MarshalIndent(resGet, "", " ")

	if strings.Contains(string(file), "\"description\": \"admob\"") {
		resGet = gojsonq.New().JSONString(string(file)).WhereNotEqual("description", "admod-optout").Get()
	}

	if string(file) == "[]" {
		jq2 := gojsonq.New().File(filePath)
		res2 := jq2.WhereNotNil("id")
		for param, value := range params {
			if strings.ToLower(param) == "platform" {
				res2 = jq2.WhereEqual(param, strings.ToLower(value[0]))
			}
		}
		resGet2 := res2.SortBy("value", "desc").Get()
		enc := json.NewEncoder(w)
		enc.SetIndent("", "	")
		enc.Encode(resGet2)
	} else {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "	")
		enc.Encode(resGet)
	}
}

func createAdNetwork(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var adNetwork adNetwork
	reqBody = []byte(strings.ToLower(string(reqBody)))
	json.Unmarshal(reqBody, &adNetwork)
	jq := gojsonq.New().File(filePath)
	res := jq.Max("id")
	adNetwork.ID = int64(res) + 1
	adNetworks = append(adNetworks, adNetwork)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "	")
	enc.Encode(adNetwork)
	file, _ := json.MarshalIndent(adNetworks, "", " ")
	_ = ioutil.WriteFile(filePath, file, 0644)
}

func deleteAdNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		fmt.Println("Query ERROR")
		return
	}

	for index, adNetwork := range adNetworks {
		if adNetwork.ID == id {
			adNetworks = append(adNetworks[:index], adNetworks[index+1:]...)

			enc := json.NewEncoder(w)
			enc.SetIndent("", "	")
			enc.Encode(adNetwork)
		}
	}

	file, _ := json.MarshalIndent(adNetworks, "", " ")
	_ = ioutil.WriteFile(filePath, file, 0644)
}

func updateAdNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		fmt.Println("Query ERROR")
		return
	}
	for index, adNetwork := range adNetworks {
		if adNetwork.ID == id {
			reqBody, _ := ioutil.ReadAll(r.Body)
			json.Unmarshal(reqBody, &adNetwork)
			adNetworks[index].Value = adNetwork.Value
			enc := json.NewEncoder(w)
			enc.SetIndent("", "	")
			enc.Encode(adNetwork)
		}
	}
	file, _ := json.MarshalIndent(adNetworks, "", " ")
	_ = ioutil.WriteFile(filePath, file, 0644)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/adnetworks", returnAllAdNetworks)
	myRouter.HandleFunc("/adnetwork", createAdNetwork).Methods("POST")
	myRouter.HandleFunc("/adnetwork/{id}", deleteAdNetwork).Methods("DELETE")
	myRouter.HandleFunc("/adnetwork/{id}", updateAdNetwork).Methods("POST")
	myRouter.HandleFunc("/adnetwork/{adtype}", returnAdType)
	myRouter.HandleFunc("/adnetwork", queryAdNetworks)
	log.Fatal(http.ListenAndServe(":8080", myRouter))
}

func main() {
	s, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	json.Unmarshal([]byte(s), &adNetworks)

	handleRequests()
}
