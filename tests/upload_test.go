package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/api"
	"backend/internal/db"
	"backend/internal/middleware"
)

// TestUploadSingleFile tests a single file upload via the upload handler.
func TestUploadSingleFile(t *testing.T) {
	fs, user, conn := SetupTest(t)

	// Build multipart body
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, err := w.CreateFormFile("myFile", "hello.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	io.Copy(part, strings.NewReader("hello world"))
	// optionally include folder_id field: w.WriteField("folder_id", "")
	w.Close()

	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Attach user in context (simulate auth)
	req = req.WithContext(middleware.WithUser(req.Context(), user))

	rr := httptest.NewRecorder()
	handler := api.NewUploadHandler(fs)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify DB: file record created
	var files []db.File
	if err := conn.Find(&files).Error; err != nil {
		t.Fatalf("db find files: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("expected at least 1 file in DB after upload")
	}
	t.Logf("found %d file(s) in DB", len(files))
}
