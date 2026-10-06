package main

import (
	"archive/zip"
	"encoding/json"
	"io"
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

func getSupportedMediaTypes() []string {
	return []string{
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/xml",
		"image/jpeg",
		"image/png",
	}
}

func main() {
	http.HandleFunc("/api/archive/files", func(w http.ResponseWriter, r *http.Request) {
		reader, readerErr := r.MultipartReader()
		if readerErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		form, formErr := reader.ReadForm(1000) // why 1000?
		if formErr != nil {
			// could not fit the non-file parts in memory
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		zw := zip.NewWriter(w)
		defer zw.Close()

		for _, headers := range form.File {
			for _, header := range headers {
				f, fErr := header.Open()
				if fErr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				defer f.Close()

				// TODO: Check for MIME type
				zf, zfErr := zw.Create(header.Filename)
				if zfErr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				_, cErr := io.Copy(zf, f)
				if cErr != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
	})

	http.HandleFunc("/api/archive/information", func(w http.ResponseWriter, r *http.Request) {
		reader, readerErr := r.MultipartReader()
		if readerErr != nil {
			// wrong content-type of a request
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		form, formErr := reader.ReadForm(1000) // why 1000?
		if formErr != nil {
			// could not fit the non-file parts in memory
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		// only one file is required
		if len(form.File) != 1 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		header := form.File["file"][0]
		if header.Header.Get("Content-Type") != "application/zip" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		file, fileError := header.Open()
		if fileError != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		zr, zrErr := zip.NewReader(file, header.Size)
		if zrErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := new(Response{
			Filename:    header.Filename,
			ArchiveSize: float64(header.Size),
		})
		var objectFiles []ObjectFile

		for _, f := range zr.File {
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

		response.Files = objectFiles
		jsonFiles, jsonErr := json.Marshal(response)
		if jsonErr == nil {
			w.Write(jsonFiles)
		}
	})

	http.ListenAndServe(":80", nil)
}
