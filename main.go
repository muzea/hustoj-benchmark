package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/muzea/measure"
)

type config struct {
	URL      string `json:"url"`
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

var extMap = map[string]string{
	".c":   "0",
	".cpp": "1",
}

func main() {
	concurrency := flag.Int("c", 10, "concurrency")
	total := flag.Int("n", 100, "total")
	configName := flag.String("config", "./config.json", "config file path")
	submitName := flag.String("code", "./1000.c", "submit code file path")
	flag.Parse()
	configFile, err := os.Open(*configName)
	defer configFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	configFileValue, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatal(err)
	}
	var data config
	json.Unmarshal(configFileValue, &data)
	m := measure.NewMeasure()
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	h := md5.New()
	io.WriteString(h, data.Password)
	client.PostForm(data.URL+"login.php", url.Values{
		"user_id":  {data.UserID},
		"password": {fmt.Sprintf("%x", h.Sum(nil))},
	})
	submitFile, err := os.Open(*submitName)
	defer submitFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	submitFileValue, err := ioutil.ReadAll(submitFile)
	if err != nil {
		log.Fatal(err)
	}
	solve := string(submitFileValue)
	count50x := 0
	count200 := 0
	countUnknown := 0
	r, _ := regexp.Compile("^\\d+$")
	submitURL := data.URL + "submit.php"
	ext := filepath.Ext(*submitName)
	language := extMap[ext]
	id := strings.TrimSuffix(filepath.Base(*submitName), ext)
	m.Stage("submit", func(runIndex int) int {
		resp, err := client.PostForm(submitURL, url.Values{
			"id":       {id},
			"language": {language},
			"source":   {solve},
		})
		if err != nil {
			fmt.Println(err)
			return 0
		}
		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			count50x++
			return 0
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			if r.MatchString(string(body[:])) {
				count200++
				return 1
			}
		}
		countUnknown++
		// fmt.Println("dump resp ", runIndex)
		// fmt.Println("[", string(body[:]), "]")
		return 0
	})
	start := time.Now()
	m.Run(*concurrency, *total)
	elapsed := time.Since(start)
	m.Print([]int{1}, false)
	fmt.Printf("total cost -> %dms\n", elapsed/time.Millisecond)
	second := float64(elapsed) / float64(time.Second)
	fmt.Printf(
		"ok %d\n50x %d\nunknown %d\nqps %f (valid value only)\n",
		count200,
		count50x,
		countUnknown,
		float64(count200)/second)
}
