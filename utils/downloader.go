package utils

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

// Used to run background BhavCopy Extraction operations
type BhvcpyExtractor struct {
	RDB      redis.Client
	holidays []time.Time
}

func NewExtractor(host string) *BhvcpyExtractor {
	resp, err := http.Get("https://zerodha.com/marketintel/holiday-calendar/?format=xml")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	rdb := *redis.NewClient(&redis.Options{Addr: host, Password: "", DB: 0})
	if resp.StatusCode != 200 {
		return &BhvcpyExtractor{
			RDB:      rdb,
			holidays: make([]time.Time, 0),
		}
	}

	// TODO: Fix XML extraction
	// decoder := xml.NewDecoder(resp.Body)
	// type Day struct {
	// 	title       string
	// 	description string
	// 	link        string
	// 	guid        string
	// 	pubDate     string
	// }
	// var days []Day
	// decoder.Decode(&days)

	var holidays []time.Time
	// for _, day := range days {
	// 	day, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", day.pubDate)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	holidays = append(holidays, day)
	// }

	return &BhvcpyExtractor{
		RDB:      rdb,
		holidays: holidays,
	}
}

// Clear out old data from redis
func (d *BhvcpyExtractor) Clear() {
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
func (d *BhvcpyExtractor) Push(values []string) {
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

func Find(slice []time.Time, val time.Time) bool {
	for _, item := range slice {
		if item.Format("01-02-2006") == val.Format("01-02-2006") {
			return true
		}
	}
	return false
}

// Download latest Bhavcopy and extract into Redis
func (d *BhvcpyExtractor) BhvcpyDownloader(date time.Time) {
	// Exit without downloading if today is a holiday, i.e. Saturday/Sunday
	switch date.Weekday() {
	case time.Saturday, time.Sunday:
		return
	default:
		if Find(d.holidays, date) {

		}
	}

	// Download Zip into 'tmp.zip' file
	file_str := "EQ" + date.Format("020106")
	url := "https://www.bseindia.com/download/BhavCopy/Equity/" + file_str + "_CSV.ZIP"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}

	temp := "tmp.zip"
	out, err := os.Create(temp)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
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

	fmt.Println("BhavCopy has been updated on", date)
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
