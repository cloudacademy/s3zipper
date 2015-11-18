package s3

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/AdRoll/goamz/aws"
	"github.com/AdRoll/goamz/s3"

	"github.com/cloudacademy/s3zipper/core"
)

type S3Bucket struct {
	aws_bucket *s3.Bucket
}

func New(accessKey, secretKey, region, bucket string) (*S3Bucket, error) {
	expiration := time.Now().Add(time.Hour * 1)
	auth, err := aws.GetAuth(accessKey, secretKey, "", expiration)
	if err != nil {
		return nil, err
	}
	return &S3Bucket{
		aws_bucket: s3.New(auth, aws.GetRegion(region)).Bucket(bucket),
	}, nil
}

func (z *S3Bucket) GetReader(path string) (rdr io.ReadCloser, err error) {
	rdr, err = z.aws_bucket.GetReader(path)
	if err != nil {
		switch t := err.(type) {
		case *s3.Error:
			if t.StatusCode == 404 {
				err = fmt.Errorf("File not found: %s", path)
			}
		default:
			err = fmt.Errorf("Error downloading \"%s\" - %s", path, err.Error())
		}
	}
	return
}

func (z *S3Bucket) List(prefix string) (files []core.ZipItem, err error) {

	res, err := z.aws_bucket.List(prefix, "", "", 1000)
	if err != nil {
		return
	}

	for _, s := range res.Contents {
		if s.Size == 0 {
			// skipping empty prefixes (folders)
			continue
		}
		key := strings.TrimPrefix(s.Key, prefix)
		filename := filepath.Base(key)
		folder := filepath.Dir(key)
		fmt.Println("folder:", folder)
		fmt.Println("filename:", filename)
		fmt.Println("key:", key)

		files = append(files, S3ZipItem{FileName: filename,
			Folder: folder,
			S3Path: s.Key,
		})
	}
	return
}

func (z *S3Bucket) CacheExists(prefix string) (bool, error) {
	return z.aws_bucket.Exists(prefix)
}

func (z *S3Bucket) CacheSignedUrl(prefix string) string {
	return z.aws_bucket.SignedURL(prefix, time.Now().Add(time.Minute))
}

type S3ZipItem struct {
	FileName string
	Folder   string
	S3Path   string
}

func (s S3ZipItem) GetFilename() string {
	return s.FileName
}
func (s S3ZipItem) GetFolder() string {
	return s.Folder
}
func (s S3ZipItem) GetPath() string {
	return s.S3Path
}
