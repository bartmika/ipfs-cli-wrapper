// Package urlkit provides utility functions for handling URL-related operations,
// such as downloading files from a given URL and saving them locally.
package urlkit

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// URLDownloader defines methods for downloading files.
type URLDownloader interface {
	DownloadFile(url, destination string) error
}

// DefaultURLKit is the default implementation of URLDownloader.
type DefaultURLKit struct{}

// DownloadFile downloads a file from the specified URL and saves it to the specified file path.
// It handles creating the destination file, making the HTTP GET request, and writing the response
// body to the file. If the HTTP response status is not OK (200), it returns an error.
//
// Parameters:
// - fromUrl (string): The URL of the file to download.
// - saveToFilepath (string): The local file path where the downloaded file should be saved.
//
// Returns:
//   - error: Returns an error if any step in the download process fails, including creating the
//     file, making the HTTP request, or writing the response body.
//
// Example:
//
//	err := DownloadFile("https://example.com/file.txt", "/local/path/to/file.txt")
//	if err != nil {
//	    log.Fatalf("Failed to download file: %v", err)
//	}
func (d *DefaultURLKit) DownloadFile(fromUrl string, saveToFilepath string) (err error) {
	// Create the file
	out, err := os.Create(saveToFilepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data from the specified URL
	resp, err := http.Get(fromUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
