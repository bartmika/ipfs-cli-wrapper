package urlkit

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func DownloadFile(fromUrl string, saveToFilepath string) (err error) {
	// Create the file
	out, err := os.Create(saveToFilepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(fromUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
