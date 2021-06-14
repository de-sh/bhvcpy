package utils

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var db *redis.Client

// Used to run background BhavCopy Extraction operations
type BhvcpyExtractor struct {
	holidays []time.Time
}

// Constructs a new BhvcpyExtractor instance by parsing infromation from the holiday-calendar.
func NewExtractor(host string) *BhvcpyExtractor {
	db = redis.NewClient(&redis.Options{Addr: host, Password: "", DB: 0})

	resp, err := http.Get("https://zerodha.com/marketintel/holiday-calendar/?format=xml")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Use regular constructor if holiday-calendar unavailable.
		return &BhvcpyExtractor{
			holidays: make([]time.Time, 0),
		}
	}

	// Structures used in holiday-calendar xml
	type Item struct {
		XMLName xml.Name `xml:"item"`
		PubDate string   `xml:"pubDate"`
	}

	type Channel struct {
		XMLName xml.Name `xml:"channel"`
		Items   []Item   `xml:"item"`
	}

	type Rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel Channel  `xml:"channel"`
	}

	// Byte stream read from response
	rssStream, _ := ioutil.ReadAll(resp.Body)

	days := Rss{}
	xml.Unmarshal(rssStream, &days)

	// Convert into easily usable date format
	var holidays []time.Time
	for _, day := range days.Channel.Items {
		day, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", day.PubDate)
		if err != nil {
			log.Fatal(err)
		}
		holidays = append(holidays, day)
	}

	return &BhvcpyExtractor{
		holidays: holidays,
	}
}

// Add new data into redis
func Push(pipe redis.Pipeliner, values []string) {
	if err := pipe.LPush(ctx, "name", values[1]).Err(); err != nil {
		log.Fatal(err)
	}

	pipe.HMSet(ctx, values[1],
		"code", values[0],
		"name", values[1],
		"open", values[4],
		"high", values[5],
		"low", values[6],
		"close", values[7],
	)
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
			return
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
	db.FlushDB(ctx)

	// Push in fresh data from CSV
	pipe := db.TxPipeline()
	for i, row := range lines {
		if i != 0 {
			Push(pipe, row)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		log.Fatal(err)
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
