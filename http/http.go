package http

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/wneessen/go-mail"
	"io"
	"log"
	"mime"
	"mysyqdev/zoodocs"
	"net/http"
	"os"
	"slices"
	"strings"
)

func Run() {
	http.HandleFunc("/api/archive/information", ParseArchive)
	http.HandleFunc("/api/archive/files", archiveFiles)
	http.HandleFunc("/api/mail/file", processMail)

	http.ListenAndServe(":80", nil)
}

func getMediaTypes() []string {
	return []string{zoodocs.Docx, zoodocs.Xml, zoodocs.Jpeg, zoodocs.Png}
}

func getContentTypes() []string {
	return []string{zoodocs.Docx, zoodocs.Pdf}
}

func processMail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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

	// get file from "file" section
	fileHeader := form.File["file"][0]
	contentType := fileHeader.Header.Get("Content-Type")
	i := strings.Index(fileHeader.Filename, ".")
	mimetype := mime.TypeByExtension(strings.ToLower(fileHeader.Filename[i:]))

	isContentTypeValid := slices.Contains(getContentTypes(), contentType)
	isFileTypeValid := slices.Contains(getContentTypes(), mimetype)
	if !isContentTypeValid || !isFileTypeValid {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("Content-Type is wrong")
		return
	}

	file, fileErr := fileHeader.Open()
	if fileErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("File error")
		return
	}
	defer file.Close()

	// parse .env file
	data, readErr := os.ReadFile(".env")
	if readErr != nil {
		return
	}
	newData := strings.TrimSpace(string(data))
	for line := range strings.SplitSeq(newData, "\n") {
		terms := strings.Split(line, "=")
		os.Setenv(terms[0], terms[1])
	}

	from := os.Getenv("EMAIL_FROM")
	recipients := strings.Split(form.Value["emails"][0], ",")

	message := mail.NewMsg()
	if err := message.From(from); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := message.To(recipients...); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}
	message.Subject("Files from Zoodocs💚")
	if err := message.AttachReader(fileHeader.Filename, file); err != nil {
		log.Fatalf("Failed to attach the file: %s", err)
	}

	client, err := mail.NewClient(
		"smtp.gmail.com",
		mail.WithTLSPortPolicy(mail.TLSMandatory),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(from),
		mail.WithPassword(os.Getenv("GMAIL_PASSWORD")),
	)
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}
	if err := client.DialAndSend(message); err != nil {
		log.Fatalf("failed to send mail: %s", err)
	}
}

func archiveFiles(w http.ResponseWriter, r *http.Request) {
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

			contentType := header.Header.Get("Content-Type")
			isContentTypeValid := slices.Contains(getMediaTypes(), contentType)
			i := strings.Index(header.Filename, ".")
			mimetype := mime.TypeByExtension(strings.ToLower(header.Filename[i:]))
			isFileTypeValid := slices.Contains(getMediaTypes(), mimetype)
			if !isContentTypeValid || !isFileTypeValid {
				w.WriteHeader(http.StatusBadRequest)
				// TODO: Parse all files, remember incorrect ones
				// and list them in the response
				fmt.Println("One of the files is of unsupported format")
				return
			}

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
}

func ParseArchive(w http.ResponseWriter, r *http.Request) {
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

	// FIX: could fail if request is of wrong format
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

	response := new(zoodocs.Archive{
		Filename:    header.Filename,
		ArchiveSize: float64(header.Size),
	})
	var objectFiles []zoodocs.Document

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		response.TotalSize += float64(f.UncompressedSize64)
		response.TotalFiles += 1

		objectFile := zoodocs.Document{
			FilePath: f.Name,
			Size:     float64(f.UncompressedSize64),
			Mimetype: getMimeType(f.Name),
		}
		objectFiles = append(objectFiles, objectFile)
	}

	response.Files = objectFiles
	jsonFiles, jsonErr := json.Marshal(response)
	if jsonErr == nil {
		w.Write(jsonFiles)
	}
}

func getMimeType(fileName string) string {
	i := strings.Index(fileName, ".")
	if i == -1 {
		// -1 means that no dot was found
		// in that case we assume that the file is binary
		// and assign a mimetype as for the
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/MIME_types/Common_types
		return "application/octet-stream"
	} else {
		return mime.TypeByExtension(strings.ToLower(fileName[i:]))
	}
}
