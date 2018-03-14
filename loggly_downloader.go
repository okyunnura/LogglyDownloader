package main

import (
	"flag"
	"fmt"
	"os"
	"encoding/json"
	"log"
	"github.com/comail/colog"
	"github.com/nanobox-io/golang-scribble"
	"time"
	"net/http"
	"io/ioutil"
	"net/url"
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

type Slack struct {
	Text        string `json:"text"`
	Username    string `json:"username"`
	IconEmoji  string `json:"icon_emoji"`
	IconUrl    string `json:"icon_url"`
	Channel     string `json:"channel"`
}

var token string
var account string
var fromDate *time.Time
var toDate *time.Time
var baseUrl string
var slackWebHookUrl string

func init() {
	colog.SetDefaultLevel(colog.LDebug)
	colog.SetMinLevel(colog.LTrace)
	colog.SetFormatter(&colog.StdFormatter{
		Colors: true,
		Flag:   log.Ldate | log.Ltime | log.Lshortfile,
	})
	colog.Register()

	var from, to string
	flag.StringVar(&token, "token", "", "loggly api token.")
	flag.StringVar(&account, "account", "", "loggly account name.")
	flag.StringVar(&from, "fromDate", "", "log date from.")
	flag.StringVar(&to, "toDate", "", "log date to.")
	flag.StringVar(&slackWebHookUrl, "webhook", "", "slack webhook url.")
	flag.Parse()

	jst, _ := time.LoadLocation("Asia/Tokyo")

	if len(from) > 0 {
		f, err := time.ParseInLocation(time.RFC3339, from, jst)
		if err != nil {
			log.Println("error: fromDate parse error.")
			log.Println(err)
		} else {
			fromDate = &f
		}
	}

	if len(to) > 0 {
		t, err := time.ParseInLocation(time.RFC3339, to, jst)
		if err != nil {
			log.Println("error: toDate parse error.")
			log.Println(err)
		} else {
			toDate = &t
		}
	}

	baseUrl = fmt.Sprintf("https://%s.loggly.com/apiv2", account)
}

func main() {
	paramError := false

	//token
	if len(token) < 1 {
		log.Println("error: token is empty.")
		paramError = true
	} else {
		log.Println("token: " + token)
	}

	//account
	if len(account) < 1 {
		log.Println("error: account is empty.")
		paramError = true
	} else {
		log.Println("account: " + account)
	}

	//fromDate
	if fromDate == nil {
		log.Println("error: fromDate is empty.")
		paramError = true
	} else {
		log.Println("fromDate: " + fromDate.String() + " (" + fromDate.UTC().String() + ")")
	}

	//toDate
	if toDate == nil {
		log.Println("error: toDate is empty.")
		paramError = true
	} else {
		log.Println("toDate: " + toDate.String() + " (" + toDate.UTC().String() + ")")
	}

	//slackWebHookUrl
	if len(slackWebHookUrl) < 1 {
		log.Println("error: slackWebHookUrl is empty.")
		paramError = true
	} else {
		log.Println("slackWebHookUrl: " + slackWebHookUrl)
	}

	//param check
	if paramError {
		log.Fatalln("error: params error.")
	}

	//TODO params
	root, _ := os.Getwd()
	path := "/tmp"
	dir := root + path

	if err := os.RemoveAll(dir); err != nil {
		log.Println("error: path dir not deleted")
		log.Fatal(err)
	}

	if err := os.MkdirAll(dir, 0777); err != nil {
		log.Println("error: path dir not created")
		log.Fatal(err)
	}

	db, err := scribble.New(dir, nil)
	if err != nil {
		log.Println("error: scribble new")
		log.Fatal(err)
	}

	from := fromDate.UTC().Format(time.RFC3339)
	until := toDate.UTC().Format(time.RFC3339)

	var searchResult SearchResult
	searchUrl := fmt.Sprintf("%s/search?q=*&from=%s&until=%s", baseUrl, from, until)

	if err := request(searchUrl, &searchResult); err != nil {
		log.Println("error: search request")
		log.Fatal(err)
	}
	log.Printf("%q\n", searchResult)

	var tagResult TagResult
	//TODO fix query
	//tagUrl := fmt.Sprintf("/fields/tag?rsid=%s", searchResult.Rsid.ID)
	tagUrl := fmt.Sprintf("%s/fields/tag?q=*&from=%s&until=%s", baseUrl, from, until)

	if err := request(tagUrl, &tagResult); err != nil {
		log.Println("error: tag request")
		log.Fatal(err)
	}
	log.Printf("%q\n", tagResult)

	log.Println("start uuid value load.")

	infoFile := dir + "/info.txt"
	value := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", "UUID", "AppVersion", "OsType", "OsVersion", "Model")
	if err := appendText(infoFile, value); err != nil {
		log.Println("error: append text")
		log.Fatalln(err)
	}
	for index, tag := range tagResult.Tag {
		uuid := tag.Term
		log.Printf("%d/%d [%s]\n", index+1, len(tagResult.Tag), uuid)

		logfile := dir + "/" + uuid + ".txt"
		query := fmt.Sprintf("tag:%s", uuid)
		size := 1000

		eventUrl := fmt.Sprintf("%s/events/iterate?q=%s&from=%s&until=%s&size=%d", baseUrl, query, from, until, size)

		for count := 1; eventUrl != ""; count++ {
			eventResult := EventResult{}
			if err := request(eventUrl, &eventResult); err != nil {
				log.Println("error: event request")
				log.Fatalln(err)
			}

			name := fmt.Sprintf("%02d", count)
			db.Write(uuid, name, eventResult)

			eventUrl = eventResult.Next
		}

		records, err := db.ReadAll(uuid)
		if err != nil {
			log.Println("error: db readall error.")
			log.Fatalln(err)
		}

		//results := []Result{}
		for recordIndex, f := range records {
			var eventResult EventResult
			if err := json.Unmarshal([]byte(f), &eventResult); err != nil {
				log.Println("error: json unmarshal error.")
				log.Fatalln(err)
			}
			for eventIndex, event := range eventResult.Events {
				if recordIndex == 0 && eventIndex == 0 {
					value := fmt.Sprintf("%s\t%s\t%s\t%s\t%s", uuid, event.Event.JSON.AppVersion, event.Event.JSON.OsType, event.Event.JSON.OsVersion, event.Event.JSON.Model)
					if err := appendText(infoFile, value); err != nil {
						log.Println("error: append text")
						log.Fatalln(err)
					}
				}
				value := fmt.Sprintf("%s\t%s", event.Event.JSON.Timestamp, event.Event.JSON.Message)
				if err := appendText(logfile, value); err != nil {
					log.Println("error: append text")
					log.Fatalln(err)
				}
			}
		}
	}

	params, _ := json.Marshal(Slack{
		"ログデータのダウンロード処理が完了したよ！",
		"gopher",
		"",
		"https://raw.githubusercontent.com/tenntenn/gopher-stickers/master/png/ok.png",
		"#notification"})

	resp, _ := http.PostForm(
		slackWebHookUrl,
		url.Values{"payload": {string(params)}},
	)

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		log.Println("error: slack notification")
		log.Fatalln(err)
	}
	defer resp.Body.Close()

}

func request(url string, result interface{}) error {
	log.Println("request: " + url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := new(http.Client)
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	return json.Unmarshal(body, &result)
}

func appendText(path, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text + "\n")
	if err != nil {
		return err
	}
	return nil
}
