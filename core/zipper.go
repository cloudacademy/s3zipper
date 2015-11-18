package core

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
)

type ZipItem interface {
	GetFilename() string
	GetFolder() string
	GetPath() string
}

type FileBrowser interface {
	List(string) ([]ZipItem, error)
	GetReader(string) (io.ReadCloser, error)
}

func Process(fb FileBrowser, w io.Writer, prefix string) (err error) {

	files, err := fb.List(prefix)

	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, file := range files {
		// Build safe file file name
		safeFileName := makeSafeFileName.ReplaceAllString(file.GetFilename(), "")
		if safeFileName == "" { // Unlikely but just in case
			safeFileName = "file"
		}

		// Read file from Reader, log any errors
		key_rdr, err := fb.GetReader(file.GetPath())
		defer key_rdr.Close()

		if err != nil {
			log.Printf(err.Error())
			continue
		}

		// Build a good path for the file within the zip
		zipPath := ""

		// Prefix folder name, if any
		if file.GetFolder() != "" {
			zipPath += file.GetFolder()
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

		wl, err := io.Copy(f, key_rdr)

		if err != nil {
			return err
		}

		fmt.Println("Zipped:", wl)
	}

	return nil
}

var makeSafeFileName = regexp.MustCompile(`[#<>:"/\|?*\\]`)
