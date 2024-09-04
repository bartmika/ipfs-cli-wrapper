package urlkit_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/bartmika/ipfs-cli-wrapper/internal/urlkit"
)

// TestDownloadFileSuccess tests the successful download of a file from a URL.
func TestDownloadFileSuccess(t *testing.T) {
	// Create a test HTTP server that returns a fixed response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Test file content"))
	}))
	defer server.Close()

	// Temporary file to save the downloaded content.
	tempFile := "testfile.txt"
	defer os.Remove(tempFile) // Clean up after the test.

	urlDownloader := &urlkit.DefaultURLKit{}

	err := urlDownloader.DownloadFile(server.URL, tempFile)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Check if the file was created.
	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// Adjust the size of the byte slice to match the expected content length.
	content := make([]byte, len("Test file content"))
	_, err = file.Read(content)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read file: %v", err)
	}

	expectedContent := "Test file content"
	if string(content) != expectedContent {
		t.Errorf("Expected file content %q, but got %q", expectedContent, string(content))
	}
}

// TestDownloadFileHTTPError tests the download function when the HTTP response status is not OK.
func TestDownloadFileHTTPError(t *testing.T) {
	// Create a test HTTP server that returns a 404 status.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	urlDownloader := &urlkit.DefaultURLKit{}

	err := urlDownloader.DownloadFile(server.URL, "should_not_exist.txt")
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}

	expectedError := "bad status: 404 Not Found"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, but got %q", expectedError, err.Error())
	}
}

// TestDownloadFileNetworkError tests the download function when there is a network error.
func TestDownloadFileNetworkError(t *testing.T) {
	invalidURL := "http://invalid-url"

	urlDownloader := &urlkit.DefaultURLKit{}

	err := urlDownloader.DownloadFile(invalidURL, "should_not_exist.txt")
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}

	var netErr *url.Error
	if !errors.As(err, &netErr) {
		t.Errorf("Expected a network error, but got: %v", err)
	}
}

// TestDownloadFileWriteError tests the download function when there is an error writing to the file.
func TestDownloadFileWriteError(t *testing.T) {
	// Create a test HTTP server that returns a fixed response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Test file content"))
	}))
	defer server.Close()

	// Attempt to write to an invalid file path.
	invalidFilePath := "/invalid_path/testfile.txt"

	urlDownloader := &urlkit.DefaultURLKit{}
	err := urlDownloader.DownloadFile(server.URL, invalidFilePath)
	if err == nil {
		t.Fatal("Expected an error, but got none")
	}
}
