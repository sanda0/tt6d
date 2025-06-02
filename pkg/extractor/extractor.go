package extractor

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// ExtractLinks extracts MP4 links from the given URL
func ExtractLinks(pageURL string) ([]string, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("page returned status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	bodyString := string(bodyBytes)

	// Check if this is a todaytvseries domain
	domainCheck := regexp.MustCompile(`todaytvseries\d*\.com`)
	if !domainCheck.MatchString(pageURL) {
		return extractGenericMP4Links(bodyString, pageURL)
	}

	return extractTVSeriesLinks(bodyString, pageURL)
}

// extractTVSeriesLinks extracts links from TodayTVSeries pages
func extractTVSeriesLinks(bodyString, pageURL string) ([]string, error) {
	// Extract available seasons
	seasonsRe := regexp.MustCompile(`Download Season ([0-9]{1,})`)
	seasonMatches := seasonsRe.FindAllStringSubmatch(bodyString, -1)

	if len(seasonMatches) == 0 {
		fmt.Println("No seasons found on this page")
		return extractGenericMP4Links(bodyString, pageURL)
	}

	// Get unique seasons
	seasonMap := make(map[string]bool)
	for _, match := range seasonMatches {
		if len(match) > 1 {
			seasonMap[match[1]] = true
		}
	}

	var seasons []string
	for season := range seasonMap {
		seasons = append(seasons, season)
	}

	// Extract title
	titleRe := regexp.MustCompile(`uk-article-title uk-badge1">(.*)</h1>`)
	titleMatch := titleRe.FindStringSubmatch(bodyString)
	title := "TV Series"
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}

	fmt.Printf("\nAvailable seasons to download in %s:\n", title)
	for _, season := range seasons {
		fmt.Printf("%s Season %s\n", title, season)
	}

	// Ask user for season selection
	fmt.Print("\nEnter the season number that you want to download: ")
	reader := bufio.NewReader(os.Stdin)
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read user input: %v", err)
	}

	userInput = strings.TrimSpace(userInput)
	selectedSeason := fmt.Sprintf("%02s", userInput)

	// Check if selected season exists
	found := false
	for _, season := range seasons {
		if season == userInput || season == selectedSeason ||
			fmt.Sprintf("%02s", season) == selectedSeason {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("season %s not available", userInput)
	}

	// Find episodes for the selected season
	episodePattern := fmt.Sprintf(`<div class="cell2">(S%sE[0-9]{1,})`, selectedSeason)
	episodeRe := regexp.MustCompile(episodePattern)
	episodeMatches := episodeRe.FindAllStringSubmatch(bodyString, -1)

	if len(episodeMatches) == 0 {
		return nil, fmt.Errorf("no episodes found for season %s", selectedSeason)
	}

	fmt.Printf("\nAvailable episodes in Season %s:\n", selectedSeason)
	var episodes []string
	for _, match := range episodeMatches {
		if len(match) > 1 {
			episodes = append(episodes, match[1])
			fmt.Println(match[1])
		}
	}

	// Extract download links for each episode
	var links []string
	for _, episode := range episodes {
		linkPatterns := []string{
			fmt.Sprintf(`%s</div><div class="cell[0-9]">[0-9]{1,} Mb</div><div class="cell[0-9]"><a href=['"]?([^'" >]+)['"]? class="hvr-icon-sink-away" target="_blank">.*</a></div>`, regexp.QuoteMeta(episode)),
			fmt.Sprintf(`%s[^<]*</div>[^<]*<div[^>]*>[^<]*[0-9]+\s*Mb[^<]*</div>[^<]*<div[^>]*><a\s+href=['"]([^'"]+)['"]`, regexp.QuoteMeta(episode)),
			fmt.Sprintf(`%s.*?href=['"]([^'"]+)['"].*?target="_blank"`, regexp.QuoteMeta(episode)),
			fmt.Sprintf(`%s.*?<a[^>]+href=['"]([^'"]+)['"]`, regexp.QuoteMeta(episode)),
		}

		found := false
		for _, pattern := range linkPatterns {
			re := regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(bodyString)
			if len(match) > 1 {
				links = append(links, match[1])
				fmt.Printf("Found link for %s: %s\n", episode, match[1])
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Warning: No download link found for episode %s\n", episode)
		}
	}

	return links, nil
}

// extractGenericMP4Links extracts MP4 links from generic web pages
func extractGenericMP4Links(bodyString, pageURL string) ([]string, error) {
	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %v", err)
	}

	var mp4Links []string
	fmt.Println("\nSearching for MP4 links using generic patterns...")

	// Multiple regex patterns to catch different types of MP4 links
	patterns := []string{
		// Direct .mp4 file links
		`href\s*=\s*['"](https?://[^'"]*\.mp4[^'"]*)['"]`,
		`href\s*=\s*['"](\.\/[^'"]*\.mp4[^'"]*)['"]`,
		`href\s*=\s*['"]([^'"]*\.mp4)['"]`,
		`(https?://[^\s'"<>]+\.mp4[^\s'"<>]*)`,

		// Download buttons and links
		`<a[^>]+href\s*=\s*['"](https?://[^'"]+)['"]\s*[^>]*target\s*=\s*['"]_blank['"][^>]*>`,
		`href\s*=\s*['"](https?://[^'"]+\.mp4[^'"]*)['"]\s+class\s*=\s*['"][^'"]*download[^'"]*['"]`,

		// Video source elements
		`<source[^>]+src\s*=\s*['"](https?://[^'"]*\.mp4[^'"]*)['"]`,
		`<video[^>]+src\s*=\s*['"](https?://[^'"]*\.mp4[^'"]*)['"]`,

		// Common download sites
		`href\s*=\s*['"](https?://[^'"]*(?:mediafire|mega|drive\.google|dropbox|onedrive)[^'"]*\.mp4[^'"]*)['"]`,

		// Fallback patterns
		`['"]([^'"]*\.mp4[^'"]*)['"]`,
		`(https?://[^\s<>"']+\.mp4)`,

		// Additional patterns for embedded players
		`data-url\s*=\s*['"](https?://[^'"]*\.mp4[^'"]*)['"]`,
		`data-video\s*=\s*['"](https?://[^'"]*\.mp4[^'"]*)['"]`,
	}

	for i, pattern := range patterns {
		fmt.Printf("Trying pattern %d...\n", i+1)
		re := regexp.MustCompile(`(?i)` + pattern) // Case insensitive
		matches := re.FindAllStringSubmatch(bodyString, -1)

		for _, match := range matches {
			if len(match) > 1 {
				link := strings.TrimSpace(match[1])

				// Skip empty links or special URIs
				if link == "" || strings.HasPrefix(link, "javascript:") || strings.HasPrefix(link, "data:") {
					continue
				}

				// Convert to absolute URL
				var absoluteURL *url.URL
				if strings.HasPrefix(link, "http") {
					absoluteURL, err = url.Parse(link)
				} else {
					absoluteURL, err = baseURL.Parse(link)
				}

				if err != nil {
					fmt.Printf("Warning: Failed to parse URL %s: %v\n", link, err)
					continue
				}

				finalURL := absoluteURL.String()
				if strings.Contains(strings.ToLower(finalURL), ".mp4") {
					// Check for duplicates
					duplicate := false
					for _, existing := range mp4Links {
						if existing == finalURL {
							duplicate = true
							break
						}
					}
					if !duplicate {
						fmt.Printf("Found MP4 link: %s\n", finalURL)
						mp4Links = append(mp4Links, finalURL)
					}
				}
			}
		}
	}

	if len(mp4Links) == 0 {
		fmt.Println("\nNo direct MP4 links found. Analyzing all links...")
		linkRe := regexp.MustCompile(`(?i)href\s*=\s*['"]([^'"]+)['"]`)
		allMatches := linkRe.FindAllStringSubmatch(bodyString, -1)
		fmt.Printf("Found %d total links to check\n", len(allMatches))

		for _, match := range allMatches {
			if len(match) > 1 {
				link := strings.TrimSpace(match[1])
				if strings.Contains(strings.ToLower(link), ".mp4") {
					var absoluteURL *url.URL
					if strings.HasPrefix(link, "http") {
						absoluteURL, err = url.Parse(link)
					} else {
						absoluteURL, err = baseURL.Parse(link)
					}

					if err != nil {
						fmt.Printf("Warning: Failed to parse URL %s: %v\n", link, err)
						continue
					}

					finalURL := absoluteURL.String()
					duplicate := false
					for _, existing := range mp4Links {
						if existing == finalURL {
							duplicate = true
							break
						}
					}
					if !duplicate {
						fmt.Printf("Found MP4 link (manual check): %s\n", finalURL)
						mp4Links = append(mp4Links, finalURL)
					}
				}
			}
		}
	}

	if len(mp4Links) == 0 {
		fmt.Println("\nNo MP4 links found in the page")
	} else {
		fmt.Printf("\nFound %d unique MP4 links\n", len(mp4Links))
	}

	return mp4Links, nil
}
