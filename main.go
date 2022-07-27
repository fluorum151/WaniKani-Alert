package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gen2brain/beeep"
)

type ReviewResponse struct {
	Data struct {
		Reviews []Item `json:"reviews"`
	} `json:"data"`
}

type Item struct {
	Time  string `json:"available_at"`
	Items []int  `json:"subject_ids"`
}

type KanjiResponse struct {
	Kanji struct {
		Kanji string `json:"characters"`
	} `json:"data"`
}

const Token = "10cadee5-b614-4f03-816a-08a6ef337e55"

func main() {
	for i := 1; i < 10; i++ {
		fmt.Println(GetKanji(i))
	}

	SendNotification()
	for {
		t := time.Now()
		if t.Minute() == 00 {
			SendNotification()
		}
		time.Sleep(1 * time.Minute)
	}
}

func SendNotification() {
	noticeTitles, noticeMessages := ScanReviews()
	for i, v := range noticeTitles {
		beeep.Notify(v, noticeMessages[i], "assets/information.png")
	}
}

func ScanReviews() ([]string, []string) {
	url := "https://api.wanikani.com/v2/summary"
	responseData := GetContent(url, true)
	var responseObject ReviewResponse
	json.Unmarshal(responseData, &responseObject)

	reviews := responseObject.Data.Reviews
	reviewsTotal := responseObject.Data.Reviews[0].Items
	var noticeMessage string
	var noticeTitles []string
	var noticeMessages []string

	if len(reviewsTotal) > 0 {
		noticeTitles = append(noticeTitles, fmt.Sprintf("You've got %v reviews!\n", len(reviewsTotal)))
		noticeMessage = "Your reviews are:"
		for _, v := range reviewsTotal {
			kanji := " " + GetKanji(v)
			noticeMessage += kanji
		}
		noticeMessages = append(noticeMessages, noticeMessage)
	}

	for i, v := range reviews[1:] {
		if len(v.Items) != 0 {
			newItems := v.Items
			noticeTitles = append(noticeTitles, fmt.Sprintf("You'll get %v new review(s) in %v hour(s)\n", len(newItems), i+1))
			noticeMessage = "Your reviews are:"
			for _, v := range newItems {
				kanji := " " + GetKanji(v)
				noticeMessage += kanji
			}
			noticeMessages = append(noticeMessages, noticeMessage)
			break
		}
	}
	return noticeTitles, noticeMessages
}

func GetKanji(id int) string {
	kanjiUrl := fmt.Sprintf("https://api.wanikani.com/v2/subjects/%v", id)
	responseData := GetContent(kanjiUrl, true)
	var responseObject KanjiResponse
	json.Unmarshal(responseData, &responseObject)

	return string(responseObject.Kanji.Kanji)
}

func GetContent(url string, authNeeded bool) []byte {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	if authNeeded {
		token := fmt.Sprintf("Bearer %v", Token)
		req.Header.Set("Authorization", token)
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	res.Body.Close()
	return responseData
}
