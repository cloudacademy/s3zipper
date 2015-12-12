package s3

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cloudacademy/s3zipper/core"
)

type S3Bucket struct {
	bucket string
	svc    *s3.S3
}

func checkCreds() (err error) {
	creds := credentials.NewEnvCredentials()
	_, err = creds.Get()
	if err == nil {
		// If ENV credentials are present, I don't need to check shared credentials on fs
		return
	}
	creds = credentials.NewSharedCredentials("", "")
	_, err = creds.Get()
	return
}

func New(region, bucket string) (*S3Bucket, error) {
	err := checkCreds()
	if err != nil {
		return nil, err
	}
	return &S3Bucket{
		bucket: bucket,
		svc:    s3.New(session.New(&aws.Config{Region: aws.String("us-west-2")})),
	}, nil
}

func (z *S3Bucket) GetReader(path string) (io.ReadCloser, error) {
	out, err := z.svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(z.bucket),
		Key:    aws.String(path),
	})
	return out.Body, err
}

func (z *S3Bucket) List(prefix string) (files []core.ZipItem, err error) {

	params := &s3.ListObjectsInput{
		Bucket:    aws.String(z.bucket), // Required
		Delimiter: aws.String(""),
		Marker:    aws.String(""),
		MaxKeys:   aws.Int64(1000),
		Prefix:    aws.String(prefix),
	}
	res, err := z.svc.ListObjects(params)
	if err != nil {
		return
	}
	for _, o := range res.Contents {
		if o.Size == aws.Int64(0) {
			// skipping empty prefixes (folders)
			continue
		}
		key := strings.TrimPrefix(*o.Key, prefix)
		filename := filepath.Base(key)
		folder := filepath.Dir(key)
		fmt.Println("folder:", folder)
		fmt.Println("filename:", filename)
		fmt.Println("key:", key)
		files = append(files, S3ZipItem{FileName: filename,
			Folder: folder,
			S3Path: *o.Key,
		})
	}
	return

}

func (z *S3Bucket) CacheExists(prefix string) bool {
	_, err := z.svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(z.bucket),
		Key:    aws.String(z.cacheNameFromPrefix(prefix)),
	})
	return err == nil
}

func (z *S3Bucket) CacheSignedUrl(prefix string) (string, error) {
	req, _ := z.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(z.bucket),
		Key:    aws.String(prefix),
	})
	return req.Presign(15 * time.Minute)
}

func (z *S3Bucket) cacheNameFromPrefix(prefix string) string {
	return fmt.Sprintf("%s.zip", prefix)
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
