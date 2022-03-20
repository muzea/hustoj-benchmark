package main

import (
	"crypto/md5"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	cr "github.com/muzea/concurrency"
	"github.com/muzea/counter"
)

type benchInfo struct {
	url         string
	user_id     string
	password    string
	total       string
	concurrency string
	problem_id  string
	answer      string
}

type wsData struct {
	Action  string            `json:"action"`
	ID      string            `json:"id"`
	Payload map[string]string `json:"payload"`
}

type benchResult struct {
	ID           string `json:"id"`
	Stage        string `json:"stage"`
	Count200     int    `json:"count200"`
	Count50x     int    `json:"count50x"`
	CountUnknown int    `json:"countUnknown"`
	Timecost     int    `json:"timecost"`
}

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleBench(ws *websocket.Conn, data wsData) {
	var bi = benchInfo{
		url:         data.Payload["url"],
		user_id:     data.Payload["user_id"],
		password:    data.Payload["password"],
		total:       data.Payload["total"],
		concurrency: data.Payload["concurrency"],
		problem_id:  data.Payload["problem_id"],
		answer:      data.Payload["answer"],
	}
	var err error
	err = ws.WriteJSON(struct {
		ID    string `json:"id"`
		Stage string `json:"stage"`
	}{
		ID:    data.ID,
		Stage: "pong",
	})
	if err != nil {
		log.Println("error write json: " + err.Error())
	}

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}
	h := md5.New()
	io.WriteString(h, bi.password)
	client.PostForm(bi.url+"login.php", url.Values{
		"user_id":  {bi.user_id},
		"password": {fmt.Sprintf("%x", h.Sum(nil))},
	})
	err = ws.WriteJSON(struct {
		ID    string `json:"id"`
		Stage string `json:"stage"`
	}{
		ID:    data.ID,
		Stage: "login_check",
	})
	if err != nil {
		log.Println("error write json: " + err.Error())
	}

	submitURL := bi.url + "submit.php"
	r, _ := regexp.Compile("^\\d+$")
	resp, err := client.PostForm(submitURL, url.Values{
		"id":       {bi.problem_id},
		"language": {"0"},
		"source":   {bi.answer},
	})
	if err != nil {
		ws.WriteJSON(struct {
			ID    string `json:"id"`
			Stage string `json:"stage"`
			Error string `json:"error"`
		}{
			ID:    data.ID,
			Stage: "check_mode",
			Error: fmt.Sprintln(err),
		})
		return
	}
	if resp.StatusCode != http.StatusOK {
		ws.WriteJSON(struct {
			ID    string `json:"id"`
			Stage string `json:"stage"`
			Error string `json:"error"`
		}{
			ID:    data.ID,
			Stage: "check_mode",
			Error: "response status code error " + string(resp.StatusCode),
		})
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if !r.MatchString(string(body[:])) {
		ws.WriteJSON(struct {
			ID    string `json:"id"`
			Stage string `json:"stage"`
			Error string `json:"error"`
		}{
			ID:    data.ID,
			Stage: "check_mode",
			Error: "response body not match " + string(body[:]),
		})
		return
	}

	ws.WriteJSON(struct {
		ID    string `json:"id"`
		Stage string `json:"stage"`
	}{
		ID:    data.ID,
		Stage: "benching",
	})

	start := time.Now()
	concurrency, err := strconv.Atoi(bi.concurrency)
	total, err := strconv.Atoi(bi.total)

	count50x := counter.NewCounter(0, total)
	count200 := counter.NewCounter(0, total)
	countUnknown := counter.NewCounter(0, total)
	var l sync.Mutex
	var updateBenchResult = func(stage string) {
		elapsed := time.Since(start).Milliseconds()
		var nextData benchResult = benchResult{
			Stage:        stage,
			ID:           data.ID,
			Count200:     count200.Value(),
			Count50x:     count50x.Value(),
			CountUnknown: countUnknown.Value(),
			Timecost:     int(elapsed),
		}
		l.Lock()
		ws.WriteJSON(nextData)
		l.Unlock()
	}

	cr.Run(func(runIndex int) {
		resp, err := client.PostForm(submitURL, url.Values{
			"id":       {bi.problem_id},
			"language": {"0"},
			"source":   {bi.answer},
		})
		if err != nil {
			fmt.Println(err)
			countUnknown.Plus(1)
			go updateBenchResult("bench_update")
			return
		}
		if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
			count50x.Plus(1)
			go updateBenchResult("bench_update")
			return
		}
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			if r.MatchString(string(body[:])) {
				count200.Plus(1)
				go updateBenchResult("bench_update")
				return
			}
		}
		countUnknown.Plus(1)
		// fmt.Println("dump resp ", runIndex)
		// fmt.Println("[", string(body[:]), "]")
		go updateBenchResult("bench_update")
		return
	}, concurrency, total)
	elapsed := time.Since(start)
	fmt.Printf("total cost -> %dms\n", elapsed/time.Millisecond)
	count50x.Flush()
	count200.Flush()
	countUnknown.Flush()
	updateBenchResult("bench_end")
}

func jsonApi(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("error get connection")
		log.Fatal(err)
	}
	defer ws.Close()
	var data wsData
	for {
		err = ws.ReadJSON(&data)
		if err != nil {
			var wsClose = "websocket: close 1001 (going away)"
			if err.Error() == wsClose {
				return
			} else {
				log.Println("error read json")
				log.Fatal(err.Error())
			}
		}

		if data.Action == "start_bench" {
			// log.Println("recive data ", data)
			handleBench(ws, data)
		}
	}
}

//go:embed build/public/*
var staticFS embed.FS

type embedFileSystem struct {
	http.FileSystem
	indexes bool
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	f, err := e.Open(path)
	if err != nil {
		return false
	}

	s, _ := f.Stat()
	if s.IsDir() && !e.indexes {
		return false
	}

	return true
}

func EmbedFolder(fsEmbed embed.FS, targetPath string, index bool) static.ServeFileSystem {
	subFS, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(subFS),
		indexes:    index,
	}
}

func websocketGin() {
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/ws", jsonApi)

	fs := EmbedFolder(staticFS, "build/public", true)
	r.Use(static.Serve("/", fs))

	r.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	r.Run(":8000")
}

func mustFS() http.FileSystem {
	sub, err := fs.Sub(staticFS, "build/public")

	if err != nil {
		panic(err)
	}

	return http.FS(sub)
}

func main() {
	mime.AddExtensionType(".js", "application/javascript")
	websocketGin()
}
