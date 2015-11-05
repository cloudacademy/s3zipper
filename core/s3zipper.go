package core

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"
)

type S3Zipper struct {
	aws_bucket *s3.Bucket
}

func New(c Configuration) (*S3Zipper, error) {
	expiration := time.Now().Add(time.Hour * 1)
	auth, err := aws.GetAuth(c.AccessKey, c.SecretKey, "", expiration)
	if err != nil {
		return nil, err
	}
	return &S3Zipper{
		aws_bucket: s3.New(auth, aws.GetRegion(c.Region)).Bucket(c.Bucket),
	}, nil
}

func (z *S3Zipper) Process(w io.Writer, prefix string) (err error) {

	files, err := z.getFilesFromS3(prefix)

	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, file := range files {
		// Build safe file file name
		safeFileName := makeSafeFileName.ReplaceAllString(file.FileName, "")
		if safeFileName == "" { // Unlikely but just in case
			safeFileName = "file"
		}

		// Read file from S3, log any errors
		rdr, err := z.aws_bucket.GetReader(file.S3Path)
		defer rdr.Close()

		if err != nil {
			switch t := err.(type) {
			case *s3.Error:
				if t.StatusCode == 404 {
					log.Printf("File not found. %s", file.S3Path)
				}
			default:
				log.Printf("Error downloading \"%s\" - %s", file.S3Path, err.Error())
			}
			continue
		}

		// Build a good path for the file within the zip
		zipPath := ""

		// Prefix folder name, if any
		if file.Folder != "" {
			zipPath += file.Folder
			if !strings.HasSuffix(zipPath, "/") {
				zipPath += "/"
			}
		}
		zipPath += safeFileName

		// We have to set a special flag so zip files recognize utf file names
		// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
		h := &zip.FileHeader{
			Name:   zipPath,
			Method: zip.Deflate,
			Flags:  0x800,
		}

		f, _ := zipWriter.CreateHeader(h)

		fmt.Println("Zipping:", zipPath)

		wl, err := io.Copy(f, rdr)

		if err != nil {
			return err
		}

		fmt.Println("Zipped:", wl)
	}

	return nil
}

func (z *S3Zipper) getFilesFromS3(prefix string) (files []*ZipItem, err error) {

	res, err := z.aws_bucket.List(prefix, "", "", 1000)
	if err != nil {
		return
	}

	for _, s := range res.Contents {
		filename := filepath.Base(s.Key)
		folder := filepath.Dir(s.Key)

		files = append(files, &ZipItem{FileName: filename,
			Folder: folder,
			S3Path: s.Key,
		})
	}
	return
}

type Configuration struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	Port      int
}

type ZipItem struct {
	FileName string
	Folder   string
	S3Path   string
}

var makeSafeFileName = regexp.MustCompile(`[#<>:"/\|?*\\]`)
