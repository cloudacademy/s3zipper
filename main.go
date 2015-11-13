package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"net/http"

	"github.com/cloudacademy/s3zipper/core"
)

var s3zipper *core.S3Zipper

var configFile string

func init() {
	flag.StringVar(&configFile, "c", "./conf.json", "config file path")
}

func main() {
	flag.Parse()

	configJSON, err := os.Open(configFile)
	checkError(err)

	decoder := json.NewDecoder(configJSON)

	config := core.Configuration{}
	err = decoder.Decode(&config)
	checkError(err)

	s3zipper, err = core.New(config)
	checkError(err)

	fmt.Println("Running on port", config.Port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get "id" URL params
	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids) < 1 {
		http.Error(w, "S3 File Zipper. Pass ?id= to use.", 500)
		return
	}
	id := ids[0]

	// Get "version" URL params
	vers, ok := r.URL.Query()["v"]
	v := "1"
	if ok && len(vers) > 0 {
		v = vers[0]
	}

	prefix := fmt.Sprintf("%s/%s", id, v)
	cache_file := fmt.Sprintf("%s.zip", prefix)

	exists, err := s3zipper.CacheExists(cache_file)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s", "File doesn't exist"), 404)
		return
	}

	if exists {
		cache_url := s3zipper.CacheSignedUrl(cache_file)
		//TODO must be converted to Permanent redirection code
		http.Redirect(w, r, cache_url, 302)
		return
	}

	// Start processing the response
	w.Header().Add("Content-Disposition", "attachment; filename=\""+prefix+"\".zip")
	w.Header().Add("Content-Type", "application/zip")

	s3zipper.Process(w, prefix)

	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

func checkError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
