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
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var temp_folder string

// Used to run background BhavCopy Extraction operations
type BhvcpyExtractor struct {
	holidays []time.Time
	db       *redis.Client
}

// Constructs a new BhvcpyExtractor instance by parsing infromation from the holiday-calendar.
func NewExtractor(db *redis.Client) *BhvcpyExtractor {
	// Set folder to be used for temporary file storage
	temp_folder = os.Getenv("BHVCPY_TEMP")

	resp, err := http.Get("https://zerodha.com/marketintel/holiday-calendar/?format=xml")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Use regular constructor if holiday-calendar unavailable.
		return &BhvcpyExtractor{
			holidays: make([]time.Time, 0),
			db:       db,
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
		db:       db,
	}
}

// Add new data into redis
func Push(pipe *redis.Pipeliner, values []string) {
	name := strings.Trim(values[1], " ")
	if err := (*pipe).SAdd(ctx, "names", name).Err(); err != nil {
		log.Fatal(err)
	}

	(*pipe).HMSet(ctx, name,
		"code", values[0],
		"name", name,
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

	// Ensure temporary folder exists or create it
	if err := os.MkdirAll(temp_folder, 0755); err != nil {
		log.Fatal(err)
	}

	temp := filepath.Join(temp_folder, "tmp.zip")
	out, err := os.Create(temp)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Extract CSV from zip file
	Unzip(temp)

	// Open CSV file
	f, err := os.Open(filepath.Join(temp_folder, file_str+".CSV"))
	if err != nil {
		log.Fatal(err)
	}

	// Parse CSV file
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Clear out old data
	d.db.FlushDB(ctx)

	// Push in fresh data from CSV
	pipe := d.db.TxPipeline()
	for i, row := range lines {
		if i != 0 {
			Push(&pipe, row)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("BhavCopy has been updated on", date)
}

// Unzip into local temporary folder
func Unzip(path string) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range reader.File {
		path := filepath.Join(temp_folder, file.Name)
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
