package zoodocs

type Document struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	Mimetype string  `json:"mimetype"`
}

type Archive struct {
	Filename    string     `json:"filename"`
	ArchiveSize float64    `json:"archive_size"`
	TotalSize   float64    `json:"total_size"`
	TotalFiles  float64    `json:"total_files"`
	Files       []Document `json:"files"`
}

// MIME file types
const (
	Docx string = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	Xml  string = "application/xml"
	Jpeg string = "image/jpeg"
	Png  string = "image/png"
	Pdf  string = "application/pdf"
)
