package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"net/http"

	"github.com/cloudacademy/s3zipper/core"
)

var s3zipper *core.S3Zipper

func main() {
	var config = core.Configuration{}
	configFile, _ := os.Open("conf.json")
	decoder := json.NewDecoder(configFile)
	err := decoder.Decode(&config)
	if err != nil {
		panic("Error reading conf")
	}

	s3zipper, err = core.New(config)
	if err != nil {
		panic("Error reading conf")
	}
	fmt.Println("Running on port", config.Port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get "ref" URL params
	refs, ok := r.URL.Query()["ref"]
	if !ok || len(refs) < 1 {
		http.Error(w, "S3 File Zipper. Pass ?ref= to use.", 500)
		return
	}
	ref := refs[0]

	// Start processing the response
	w.Header().Add("Content-Disposition", "attachment; filename=\""+ref+".zip\"")
	w.Header().Add("Content-Type", "application/zip")

	s3zipper.Process(w, ref)

	log.Printf("%s\t%s\t%s", r.Method, r.RequestURI, time.Since(start))
}
