# Zoodocs
Take home assignment for [Doodocs internship](https://github.com/doodocs/doodocs-days/blob/main/backend/README.md)

## How to run
1. Clone this repo
2. Create a `.env` file and put two variables there:
```
EMAIL_FROM="youremail@gmail.com"
GMAIL_PASSWORD="app password for you gmail account"
```
2. run `go run main.go`

## On mail sending
I wanted to use zero external dependencies, only std library. However, turns out net/smtp does not support file attachments. Writing it myself from scratch would take time and it would not guarantee stability and things like that, so in the end I am using a library.


## Manual Testing
I tested my code using cURL. Here are the commands:

### Task 1
curl -X POST \
-F "file=@/path/to/file/Archive.zip;type=application/zip" \
http://localhost:80/api/archive/information

### Task 2
curl -X POST \
-F "files[]=@/path/to/file/documents.docx" \
-F "files[]=@/path/to/file/image.JPG" \
-o test.zip \
http://localhost:80/api/archive/files

### Task 3
curl -X POST \
-F file="@/path/to/file/test.docx;type=application/vnd.openxmlformats-officedocument.wordprocessingml.document" \
-F emails="doodocs@doodocs-email.kz,noreply@doodocs-email.kz" \
http://localhost:80/api/mail/file
