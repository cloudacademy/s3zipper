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
	"github.com/dgrijalva/jwt-go"
)

var s3bucket *s3.S3Bucket

var configFile string
var conf *Configuration

type Configuration struct {
	Bucket       string
	Region       string
	Port         int
	JWTSharedKey string
}

func init() {
	flag.StringVar(&configFile, "c", "./conf.json", "config file path")
}

func main() {
	flag.Parse()

	configJSON, err := os.Open(configFile)
	checkError(err)

	decoder := json.NewDecoder(configJSON)

	conf = &Configuration{}
	err = decoder.Decode(conf)
	checkError(err)

	s3bucket, err = s3.New(conf.Region, conf.Bucket)
	checkError(err)

	fmt.Println("Running on port", conf.Port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+strconv.Itoa(conf.Port), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get "id" URL params
	ids, ok := r.URL.Query()["id"]
	if !ok || len(ids) < 1 {
		http.Error(w, "S3 File Zipper. Pass JWT token with ?id=", 400)
		return
	}
	prefix, err := parseJWT(ids[0])
	if err != nil {
		log.Printf("Error decoding JWT: %s\n", err)
		http.Error(w, "Invalid JWT signature", 400)
		return
	}

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

func parseJWT(tokenString string) (string, error) {

	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(conf.JWTSharedKey), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token: %v", token.Raw)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("claims not a valid map")
	}

	id, ok := claims["id"]
	if !ok {
		return "", fmt.Errorf("missing `id` property in jwt token")
	}

	return id.(string), nil
}
