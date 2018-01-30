package main

import (
	"flag"
	"fmt"
	"os"
)

type Result struct {
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

func init() {
	flag.StringVar(&token, "token", "", "loggly api token.")
	flag.Parse()
}

func main() {
	if len(token) < 1 {
		fmt.Println("token is empty.")
		os.Exit(0)
	}

	fmt.Println("token:", token)

}
