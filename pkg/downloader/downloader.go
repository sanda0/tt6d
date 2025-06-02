package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"tt6d/pkg/progress"
)

// Download downloads multiple files concurrently
func Download(links []string, downloadFolder string, concurrentDownloads int) error {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrentDownloads)

	for i, link := range links {
		wg.Add(1)
		go func(index int, mp4URL string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := downloadFile(mp4URL, downloadFolder, index+1, len(links)); err != nil {
				fmt.Printf("\n[%d/%d] Error downloading %s: %v\n", index+1, len(links), mp4URL, err)
			}
		}(i, link)
	}

	wg.Wait()
	return nil
}

// downloadFile downloads a single file with progress tracking
func downloadFile(fileURL, downloadFolder string, index, totalFiles int) error {
	// Get the file
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("file download returned status code: %d", resp.StatusCode)
	}

	// Extract filename from URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return fmt.Errorf("failed to parse file URL: %v", err)
	}

	filename := path.Base(parsedURL.Path)
	if filename == "." || filename == "/" {
		filename = "video.mp4" // fallback filename
	}

	// Ensure filename ends with .mp4
	if !strings.HasSuffix(strings.ToLower(filename), ".mp4") {
		filename += ".mp4"
	}

	// Create full file path
	filePath := filepath.Join(downloadFolder, filename)

	// Handle duplicate filenames
	counter := 1
	originalPath := filePath
	for {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		// File exists, create a new name
		ext := filepath.Ext(originalPath)
		base := strings.TrimSuffix(originalPath, ext)
		filePath = fmt.Sprintf("%s_%d%s", base, counter, ext)
		counter++
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	// Get content length for progress tracking
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		contentLength = 0 // Unknown size
	}

	// Create progress writer
	progressWriter := progress.New(out, contentLength, filename, index, totalFiles)

	// Copy the response body to file with progress tracking
	_, err = io.Copy(progressWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	return nil
}
