package main

import (
	"fmt"
	"os"
	"strconv"

	"tt6d/pkg/downloader"
	"tt6d/pkg/extractor"
	"tt6d/pkg/ui"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("TT6D - TodayTVSeries6 Downloader")
		fmt.Println("Usage: tt6d <webpage_url> <download_folder> [concurrent_downloads]")
		fmt.Println("Example:")
		fmt.Println("  tt6d https://todaytvseries6.com/series/example /home/user/downloads")
		fmt.Println("  tt6d https://todaytvseries6.com/series/example /home/user/downloads 3")
		os.Exit(1)
	}

	pageURL := os.Args[1]
	downloadFolder := os.Args[2]

	// Set concurrent downloads (default: 1 for sequential downloads)
	concurrentDownloads := 1
	if len(os.Args) > 3 {
		if n, err := strconv.Atoi(os.Args[3]); err == nil && n > 0 {
			concurrentDownloads = n
		}
	}

	// Create download folder if it doesn't exist
	if err := os.MkdirAll(downloadFolder, 0755); err != nil {
		fmt.Printf("Error creating download folder: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching page: %s\n", pageURL)
	links, seriesInfo, err := extractor.ExtractContent(pageURL)
	if err != nil {
		fmt.Printf("Error extracting content: %v\n", err)
		os.Exit(1)
	}

	var selectedLinks []string
	if seriesInfo != nil {
		fmt.Printf("Found TV Series: %s\n", seriesInfo.Title)
		selectedLinks, err = ui.SelectTVSeriesEpisodes(seriesInfo)
	} else {
		if len(links) == 0 {
			fmt.Println("No MP4 links found on the page")
			return
		}
		fmt.Printf("Found %d MP4 links\n", len(links))
		selectedLinks, err = ui.GetSelectedLinks(links)
	}

	if err != nil {
		if err.Error() == "no files selected" || err.Error() == "no episodes selected" {
			fmt.Println("\nNo files selected for download")
			return
		}
		fmt.Printf("Error in link selection: %v\n", err)
		os.Exit(1)
	}

	// Download selected files using the downloader package
	if err := downloader.Download(selectedLinks, downloadFolder, concurrentDownloads); err != nil {
		fmt.Printf("Error during download: %v\n", err)
		os.Exit(1)
	}

}
