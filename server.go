package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/de-sh/bhvcpy/utils"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
)

var ctx = context.Background()

// Connect to redis instance
var conn = redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})

func queryRedis(w http.ResponseWriter, r *http.Request) {
	// Parse JSON from POST for key
	decoder := json.NewDecoder(r.Body)

	type Data struct {
		Key string
	}

	var d Data
	err := decoder.Decode(&d)
	if err != nil {
		panic(err)
	}
	key := strings.ToUpper(d.Key)

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
	// Create a new BhavCopy extraction cron job to run at 6PM Mon-Fri
	c := cron.New()
	d := utils.NewExtractor(conn)
	c.AddFunc("0 18 1-5 * *", func() {
		d.BhvcpyDownloader(time.Now())
	})
	c.Start()
	defer c.Stop()

	// Add handlers to serve `/` and `/get`
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.ServeFile(w, r, "./static/index.html")
		case "POST":
			queryRedis(w, r)
		}

	})

	fmt.Println("Server listening on port 3000")

	log.Panic(
		http.ListenAndServe(":3000", nil),
	)
}
