package main

import (
	"archive/zip"
	"encoding/json"
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

		reader, readerErr := r.MultipartReader()
		if readerErr != nil {
			log.Fatal(readerErr)
		}

		form, formErr := reader.ReadForm(1000) // why 1000?
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

				for _, f := range r.File {
					if f.FileInfo().IsDir() {
						continue
					}

					response.TotalSize += float64(f.UncompressedSize64)
					response.TotalFiles += 1

					i := strings.Index(f.Name, ".")
					var mimetype string
					if i == -1 {
						// -1 means that no dot was found
						// in that case we assume that the file is binary
						// and assign a mimetype as for the
						// https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/MIME_types/Common_types
						mimetype = "application/octet-stream"
					} else {
						mimetype = mime.TypeByExtension(strings.ToLower(f.Name[i:]))
					}

					objectFile := ObjectFile{
						FilePath: f.Name,
						Size:     float64(f.UncompressedSize64),
						Mimetype: mimetype,
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
