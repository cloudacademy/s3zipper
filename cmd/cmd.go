package main

import (
	"encoding/json"
	"os"

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

	f, err := os.Create("/tmp/dat2")
	if err != nil {
		panic(err)
	}

	s3zipper.Process(f, "CloudAcademy-AWS-Certified-Developer-Associate-Level")
}
