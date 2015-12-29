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

	zipper "github.com/cloudacademy/s3zipper/core"
	"github.com/cloudacademy/s3zipper/s3"
)

var s3bucket *s3.S3Bucket

var configFile string

type Configuration struct {
	Bucket string
	Region string
	Port   int
}

func init() {
	flag.StringVar(&configFile, "c", "./conf.json", "config file path")
}

func main() {
	flag.Parse()

	configJSON, err := os.Open(configFile)
	checkError(err)

	decoder := json.NewDecoder(configJSON)

	c := Configuration{}
	err = decoder.Decode(&c)
	checkError(err)

	s3bucket, err = s3.New(c.Region, c.Bucket)
	checkError(err)

	fmt.Println("Running on port", c.Port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(c.Port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get "id" URL params
	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids) < 1 {
		http.Error(w, "S3 File Zipper. Pass ?id= to use.", 400)
		return
	}
	prefix := ids[0]

	zipname := prefix
	if name, ok := r.URL.Query()["name"]; ok && len(name) >= 1 {
		zipname = name[0]
	}

	exists := s3bucket.CacheExists(prefix)

	if exists {
		cache_url, err := s3bucket.CacheSignedUrl(prefix)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
		//TODO must be converted to Permanent redirection code
		http.Redirect(w, r, cache_url, 302)
		return
	}

	// Start processing the response
	w.Header().Add("Content-Disposition", "attachment; filename=\""+zipname+".zip\"")
	w.Header().Add("Content-Type", "application/zip")

	zipper.Process(s3bucket, w, prefix)

	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}

func checkError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
