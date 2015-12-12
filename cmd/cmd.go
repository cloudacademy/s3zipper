package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	zipper "github.com/cloudacademy/s3zipper/core"
	"github.com/cloudacademy/s3zipper/fs"
	"github.com/cloudacademy/s3zipper/s3"
)

var filesystem *fs.Filesystem
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

	if len(flag.Args()) < 2 {
		log.Fatal("Missing params: <prefix> and <output>")
	}

	prefix := flag.Arg(0)
	output := flag.Arg(1)

	configJSON, err := os.Open(configFile)
	checkError(err)

	decoder := json.NewDecoder(configJSON)

	c := Configuration{}
	err = decoder.Decode(&c)
	checkError(err)

	s3bucket, err := s3.New(c.Region, c.Bucket)
	checkError(err)
	fmt.Println(s3bucket)
	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	err = zipper.Process(s3bucket, f, prefix)
	checkError(err)

}

func checkError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
