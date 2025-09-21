package api

import (
	// "context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	// "os/user"

	"backend/internal/middleware"
	"backend/internal/services"
)

// Response model for each file
type UploadResult struct {
	FileID   string `json:"file_id,omitempty"`
	FileName string `json:"file_name,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Hash     string `json:"sha256,omitempty"`
	Error    string `json:"error,omitempty"`
}

// New Upload Handler return an handler function bound to the file service
func NewUploadHandler(fs *services.FileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		//get use from context
		user := middleware.GetUser(ctx)
		if user == nil {
			http.Error(w, "user not found in context(not authenticated)", http.StatusUnauthorized)
		}
		userID := user.ID

		//safety cap for transport layer where parsing happens
		if err := r.ParseMultipartForm(100 << 20); err != nil { // parsing limit 100MB
			http.Error(w, "no files uploaded or parse error", http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["myFile"]
		if len(files) == 0 {
			http.Error(w, "no files uploaded ", http.StatusBadRequest)
			return
		}

		results := make([]UploadResult, 0, len(files))

		for _, fh := range files {
			//Open part
			part, err := fh.Open()
			if err != nil {
				results = append(results, UploadResult{FileName: fh.Filename, Error: "open error: " + err.Error()})
				continue
			}

			// create temp file
			tmp, err := os.CreateTemp("", "upload-*.")
			if err != nil {
				part.Close()
				results = append(results, UploadResult{FileName: fh.Filename, Error: "temp file error:" + err.Error()})
				continue
			}
			tmpPath := tmp.Name()

			// Reading of the first 512 bytes to detect mime type and write them to temp + hash
			header := make([]byte, 512)
			n, _ := part.Read(header)
			mimeType := http.DetectContentType(header[:n])

			//Writer header bytes to temp and start hashing
			hasher := sha256.New()
			if n > 0 {
				if _, err := tmp.Write(header[:n]); err != nil {
					part.Close()
					tmp.Close()
					os.Remove(tmpPath)
					results = append(results, UploadResult{FileName: fh.Filename, Error: "writing in temp error:" + err.Error()})
					continue
				}

				if _, err := hasher.Write(header[:n]); err != nil {
					part.Close()
					tmp.Close()
					os.Remove(tmpPath)
					results = append(results, UploadResult{FileName: fh.Filename, Error: "hashing error:" + err.Error()})
					continue
				}
			}

			// Copy remainder of the file to temp while hashing
			written, err := io.Copy(io.MultiWriter(tmp, hasher), part)
			if err != nil {
				part.Close()
				tmp.Close()
				os.Remove(tmpPath)
				results = append(results, UploadResult{FileName: fh.Filename, Error: "Remaining file copy error:" + err.Error()})
				continue
			}

			totalSize := int64(n) + written
			sha := fmt.Sprintf("%x", hasher.Sum(nil))

			// Close reader writter before passing to service
			part.Close()
			tmp.Close()

			// every mime type is allowed

			//Call file service to process the upload
			fileID, err := fs.ProcessUpload(ctx, userID, fh.Filename, tmpPath, totalSize, mimeType, sha)

			// remove temp file regardless of success or failure
			os.Remove(tmpPath)

			if err != nil {
				results = append(results, UploadResult{FileName: fh.Filename, Size: totalSize, Hash: sha, Error: "file service error:" + err.Error()})
				continue
			}
			results = append(results, UploadResult{FileID: fileID, FileName: fh.Filename, Size: totalSize, Hash: sha})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
