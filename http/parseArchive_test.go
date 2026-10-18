package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"mysyqdev/zoodocs"
	msqhttp "mysyqdev/zoodocs/http"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"testing"
)

func TestParseArchive(t *testing.T) {
	t.Run("parsing an archive", func(t *testing.T) {
		buf := &bytes.Buffer{}
		mpw := multipart.NewWriter(buf)
		nameWriter, err := mpw.CreateFormField("file")
		if err != nil {
			fmt.Println("Could not create a form filed")
		}
		_, err = nameWriter.Write([]byte("file"))
		if err != nil {
			fmt.Println("Could not write to name")
		}

		f, err := os.Open("Archive.zip")
		if err != nil {
			fmt.Println("Could not open the file")
		}
		defer f.Close()

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", multipart.FileContentDisposition("file", "Archive.zip"))
		h.Set("Content-Type", "application/zip")
		fWriter, err := mpw.CreatePart(h)
		if err != nil {
			fmt.Println("Could not create a form file")
		}

		io.Copy(fWriter, f)
		mpw.Close()

		request, _ := http.NewRequest(http.MethodPost, "/api/archive/information", buf)
		request.Header.Add("Content-Type", mpw.FormDataContentType())

		response := httptest.NewRecorder()

		msqhttp.ParseArchive(response, request)

		got := response.Body.String()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("could not read the body")
		}
		var jsonResponse zoodocs.Archive
		err = json.Unmarshal(body, &jsonResponse)
		if err != nil {
			fmt.Println("Could not unmarshall the json")
		}
		want := "Archive.zip"

		if jsonResponse.Filename != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
