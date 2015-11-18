package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	zipper "github.com/cloudacademy/s3zipper/core"
	"github.com/cloudacademy/s3zipper/fs"
)

var filesystem *fs.Filesystem
var configFile string

type Configuration struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Port      int
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

	filesystem = fs.New()

	f, err := os.Create(output)
	if err != nil {
		panic(err)
	}
	zipper.Process(filesystem, f, prefix)
	//zipper.Process(s3bucket, f, prefix)

}

func checkError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
