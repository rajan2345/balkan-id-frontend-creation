package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/api"
	"backend/internal/db"
	"backend/internal/middleware"
)

// TestDedup verifies uploading the same content twice produces a single File but two UserFile links (ref count increased).
func TestDedup(t *testing.T) {
	fs, user, conn := SetupTest(t)

	upload := func() {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		part, _ := w.CreateFormFile("myFile", "dup.txt")
		io.Copy(part, strings.NewReader("same content"))
		w.Close()

		req := httptest.NewRequest("POST", "/upload", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
		req = req.WithContext(middleware.WithUser(req.Context(), user))

		rr := httptest.NewRecorder()
		handler := api.NewUploadHandler(fs)
		handler.ServeHTTP(rr, req)
		if rr.Code != 200 {
			t.Fatalf("upload failed: %d %s", rr.Code, rr.Body.String())
		}
	}

	// First upload
	upload()

	// Second upload (same content)
	upload()

	// Check files table has exactly one file (deduped)
	var files []db.File
	if err := conn.Find(&files).Error; err != nil {
		t.Fatalf("db find files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 deduped file, found %d", len(files))
	}

	// Check user_files references count (should be 2)
	var refs []db.UserFile
	if err := conn.Where("file_id = ?", files[0].ID).Find(&refs).Error; err != nil {
		t.Fatalf("db find refs: %v", err)
	}
	if len(refs) < 2 {
		t.Fatalf("expected at least 2 user-file references, got %d", len(refs))
	}
}
