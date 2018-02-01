package main

import (
	"flag"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"log"
)

type SearchResult struct {
	Rsid struct {
		Status      string  `json:"status"`
		DateFrom    int64   `json:"date_from"`
		ElapsedTime float64 `json:"elapsed_time"`
		DateTo      int64   `json:"date_to"`
		ID          string  `json:"id"`
	} `json:"rsid"`
}

type TagResult struct {
	TotalEvents int `json:"total_events"`
	Tag []struct {
		Count int    `json:"count"`
		Term  string `json:"term"`
	} `json:"tag"`
	UniqueFieldCount int `json:"unique_field_count"`
}

type EventResult struct {
	Events []struct {
		Raw       interface{} `json:"raw"`
		Logtypes  []string    `json:"logtypes"`
		Timestamp int64       `json:"timestamp"`
		Unparsed  interface{} `json:"unparsed"`
		Logmsg    interface{} `json:"logmsg"`
		ID        string      `json:"id"`
		Tags      []string    `json:"tags"`
		Event struct {
			JSON struct {
				Model      string      `json:"model"`
				Level      string      `json:"level"`
				Timestamp  interface{} `json:"timestamp"`
				OsVersion  string      `json:"os_version"`
				Tag        string      `json:"tag"`
				Message    string      `json:"message"`
				AppVersion string      `json:"app_version"`
				OsType     string      `json:"os_type"`
			} `json:"json"`
			HTTP struct {
				ClientHost  string `json:"clientHost"`
				ContentType string `json:"contentType"`
			} `json:"http"`
		} `json:"event"`
	} `json:"events"`
	Next string `json:"next"`
}

var token string

var account string

var baseUrl string

func init() {
	flag.StringVar(&token, "token", "", "loggly api token.")
	flag.StringVar(&account, "account", "", "loggly account name.")
	flag.Parse()

	baseUrl = fmt.Sprintf("https://%s.loggly.com/apiv2", account)
}

func main() {
	if len(token) < 1 {
		fmt.Println("token is empty.")
		os.Exit(0)
	}
	fmt.Println("token:" + token)

	if len(account) < 1 {
		fmt.Println("account is empty.")
		os.Exit(0)
	}
	fmt.Println("account:" + account)

	// TODO params
	day := "01-30"
	from := fmt.Sprintf("2018-%sT01:00:00.000Z", day)
	until := fmt.Sprintf("2018-%sT09:00:00.000Z", day)

	var searchResult SearchResult
	searchUrl := fmt.Sprintf("/search?q=*&from=%s&until=%s", from, until)

	if err := request(searchUrl, &searchResult); err != nil {
		fmt.Println("error:search request")
		log.Fatal(err)
		os.Exit(0)
	}
	fmt.Printf("%q\n", searchResult)

	var tagResult TagResult
	tagUrl := fmt.Sprintf("/fields/tag?rsid=%s", searchResult.Rsid.ID)

	if err := request(tagUrl, &tagResult); err != nil {
		fmt.Println("error:tag request")
		log.Fatal(err)
		os.Exit(0)
	}
	fmt.Printf("%q\n", tagResult)

}

func request(path string, result interface{}) error {
	fmt.Println("request:" + baseUrl + path)
	req, _ := http.NewRequest("GET", baseUrl+path, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := new(http.Client)
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(body, &result)
}
