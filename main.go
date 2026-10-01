package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	// route 1
	http.HandleFunc("/api/archive/information", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received request")

		reader, readerErr := r.MultipartReader()
		if readerErr != nil {
			log.Fatal(readerErr)
		}

		form, formErr := reader.ReadForm(1000)
		if formErr != nil {
			log.Fatal(formErr)
		}

		for _, headers := range form.File {
			for _, header := range headers {
				file, _ := header.Open()
				fmt.Println(file)
				r, _ := zip.NewReader(file, header.Size)
				for _, f := range r.File {
					fmt.Printf("Contents of: %s:\n", f.Name)
					rc, err := f.Open()
					if err != nil {
						log.Fatal(err)
					}
					_, _ = io.CopyN(os.Stdout, rc, 68)
					rc.Close()
					fmt.Println()
				}
			}
		}
	})

	http.ListenAndServe(":80", nil)
}
