package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strings"
)

type ObjectFile struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	Mimetype string  `json:"mimetype"`
}

type Response struct {
	Filename    string       `json:"filename"`
	ArchiveSize float64      `json:"archive_size"`
	TotalSize   float64      `json:"total_size"`
	TotalFiles  float64      `json:"total_files"`
	Files       []ObjectFile `json:"files"`
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

		var response Response
		var objectFiles []ObjectFile

		for _, headers := range form.File {
			for _, header := range headers {
				response.Filename = header.Filename
				response.ArchiveSize = float64(header.Size)

				file, _ := header.Open()
				r, _ := zip.NewReader(file, header.Size)
				// TODO: Check for directories
				for _, f := range r.File {
					response.TotalSize += float64(f.UncompressedSize64)
					response.TotalFiles += 1

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
		response.Files = objectFiles

		jsonFiles, jsonErr := json.Marshal(response)
		if jsonErr == nil {
			w.Write(jsonFiles)
		}
	})

	http.ListenAndServe(":80", nil)
}
