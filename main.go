package main

import (
	"archive/zip"
	"strings"
	// "encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
)

type ObjectFile struct {
	FilePath string
	Size     float64
	Mimetype string
}

type Response struct {
	Filename    string
	ArchiveSize float64
	TotalSize   float64
	TotalFiles  float64
	Files       []ObjectFile
}

func main() {
	http.HandleFunc("/api/archive/information", func(w http.ResponseWriter, r *http.Request) {
		// multipart/form-data
		// section name: file
		// if not archive/zip: return error

		fmt.Println("Received request")

		reader, readerErr := r.MultipartReader()
		if readerErr != nil {
			log.Fatal(readerErr)
		}

		form, formErr := reader.ReadForm(1000)
		if formErr != nil {
			log.Fatal(formErr)
		}

		var objectFiles []ObjectFile
		for _, headers := range form.File {
			for _, header := range headers {
				file, _ := header.Open()
				r, _ := zip.NewReader(file, header.Size)
				for _, f := range r.File {
					i := strings.Index(f.Name, ".")
					objectFile := ObjectFile{
						FilePath: f.Name,
						Size:     float64(f.UncompressedSize64),
						Mimetype: mime.TypeByExtension(strings.ToLower(f.Name[i:])),
					}
					objectFiles = append(objectFiles, objectFile)
				}
				file.Close()
			}
		}

		fmt.Println(objectFiles)
	})

	http.ListenAndServe(":80", nil)
}
