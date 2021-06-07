package main

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
)

var ctx = context.Background()

// Used to run multiple client operations
type DownloadCSV struct {
	RDB redis.Client
}

func NewDownloader(host string) *DownloadCSV {
	return &DownloadCSV{
		RDB: *redis.NewClient(&redis.Options{Addr: host, Password: "", DB: 0}),
	}
}

// Clear out old data from redis
func (d *DownloadCSV) Clear() {
	var names []string
	len, err := d.RDB.LLen(ctx, "name").Result()
	if err != nil {
		log.Fatal(err)
	}

	for i := int64(0); i < len; i++ {
		name, err := d.RDB.LIndex(ctx, "name", i).Result()
		if err != nil {
			log.Fatal(err)
		}

		names = append(names, name)
	}

	if err := d.RDB.Del(ctx, "name").Err(); err != nil {
		log.Fatal(err)
	}

	for _, name := range names {
		if err := d.RDB.Del(ctx, name).Err(); err != nil {
			log.Fatal(err)
		}
	}
}

// Add new data into redis
func (d *DownloadCSV) Push(values []string) {
	if err := d.RDB.LPush(ctx, "name", values[1]).Err(); err != nil {
		log.Fatal(err)
	}

	if err := d.RDB.HMSet(ctx, values[1],
		"code", values[0],
		"name", values[1],
		"open", values[4],
		"high", values[5],
		"low", values[6],
		"close", values[7],
	).Err(); err != nil {
		log.Fatal(err)
	}
}

// Download latest Bhavcopy and extract into Redis
func (d *DownloadCSV) BhvcpyDownloader(date time.Time) {
	// Exit without downloading if today is a holiday, i.e. Saturday/Sunday
	switch date.Weekday() {
	case time.Saturday, time.Sunday:
		return
	}

	// Download Zip into 'tmp.zip' file
	date_str := date.Format("02012006")
	file_str := "EQ" + date_str
	url := "https://www.bseindia.com/download/BhavCopy/Equity/" + file_str + ".ZIP"
	r, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return
	}

	temp := "tmp.zip"
	out, err := os.Create(temp)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, r.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Extract 'tmp.zip'
	Unzip(temp)

	// Open CSV file
	f, err := os.Open("./tmp/" + file_str + ".CSV")
	if err != nil {
		log.Fatal(err)
	}

	// Parse CSV file
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Clear out old data
	d.Clear()

	// Push in fresh data from CSV
	for i, row := range lines {
		if i != 0 {
			d.Push(row)
		}
	}
}

// Unzip into local `./tmp`
func Unzip(path string) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll("./tmp", 0755); err != nil {
		log.Fatal(err)
	}

	for _, file := range reader.File {
		path := filepath.Join("./tmp", file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			log.Fatal(err)
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			log.Fatal(err)
		}
	}
}

func queryRedis(w http.ResponseWriter, r *http.Request) {
	key := strings.ToUpper(r.URL.Query().Get("key"))
	// Connect to redis instance
	conn := redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})

	// Find length of list
	nos, err := conn.LLen(ctx, "name").Result()
	if err != nil {
		log.Fatal(err)
	}

	// Recieve names
	var names []string
	for code := nos - 1; code >= 0; code-- {
		name, err := conn.LIndex(ctx, "name", code).Result()
		if err != nil {
			log.Fatal(err)
		} else {
			if strings.Contains(name, key) {
				names = append(names, name)
			}
		}
	}

	// Recieve entries associated with selected names
	var entries []map[string]string
	for _, name := range names {
		entry, err := conn.HGetAll(ctx, name).Result()
		if err != nil {
			log.Fatal(err)
		}
		entries = append(entries, entry)
	}

	// Return a json output
	out := make(map[string][]map[string]string)
	out["entries"] = entries
	jsonString, _ := json.Marshal(out)

	w.Write([]byte(jsonString))
}

func main() {
	// Create a new BhavCopy downloader cron job to run at 6PM Mon-Fri
	c := cron.New()
	d := NewDownloader("localhost:6739")
	c.AddFunc("0 18 1-5 * *", func() {
		d.BhvcpyDownloader(time.Now())
	})
	c.Start()
	defer c.Stop()

	// Add handlers to serve `/` and `/get`
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/get", queryRedis)
	fmt.Println("Server listening on port 3000")

	log.Panic(
		http.ListenAndServe(":3000", nil),
	)
}
