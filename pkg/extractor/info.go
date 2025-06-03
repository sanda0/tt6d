package extractor

import (
	"fmt"
	"regexp"
)

// TVSeriesInfo contains information about available TV series seasons and episodes
type TVSeriesInfo struct {
	Title   string
	Seasons map[string][]Episode
}

// Episode represents a single episode with its links
type Episode struct {
	ID    string // e.g., "S01E01"
	Links []string
}

// ExtractTVSeriesInfo extracts TV series information without prompting for selection
func ExtractTVSeriesInfo(bodyString string) (*TVSeriesInfo, error) {
	// Extract available seasons
	seasonsRe := regexp.MustCompile(`Download Season ([0-9]{1,})`)
	seasonMatches := seasonsRe.FindAllStringSubmatch(bodyString, -1)

	if len(seasonMatches) == 0 {
		return nil, fmt.Errorf("no seasons found")
	}

	// Extract title
	titleRe := regexp.MustCompile(`uk-article-title uk-badge1">(.*)</h1>`)
	titleMatch := titleRe.FindStringSubmatch(bodyString)
	title := "TV Series"
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}

	info := &TVSeriesInfo{
		Title:   title,
		Seasons: make(map[string][]Episode),
	}

	// Get unique seasons
	seasonMap := make(map[string]bool)
	for _, match := range seasonMatches {
		if len(match) > 1 {
			seasonMap[match[1]] = true
		}
	}

	// Process each season
	for season := range seasonMap {
		seasonNum := fmt.Sprintf("%02s", season)
		episodePattern := fmt.Sprintf(`<div class="cell2">(S%sE[0-9]{1,})`, seasonNum)
		episodeRe := regexp.MustCompile(episodePattern)
		episodeMatches := episodeRe.FindAllStringSubmatch(bodyString, -1)

		// Use a map to deduplicate episodes
		episodeMap := make(map[string]Episode)
		for _, match := range episodeMatches {
			if len(match) < 2 {
				continue
			}

			epID := match[1]
			if _, exists := episodeMap[epID]; !exists {
				episodeMap[epID] = Episode{ID: epID}
			}

			// Find download links for this episode
			linkPatterns := []string{
				fmt.Sprintf(`%s</div><div class="cell[0-9]">[0-9]{1,} Mb</div><div class="cell[0-9]"><a href=['"]?([^'" >]+)['"]? class="hvr-icon-sink-away" target="_blank">.*</a></div>`, regexp.QuoteMeta(match[1])),
				fmt.Sprintf(`%s[^<]*</div>[^<]*<div[^>]*>[^<]*[0-9]+\s*Mb[^<]*</div>[^<]*<div[^>]*><a\s+href=['"]([^'"]+)['"]`, regexp.QuoteMeta(match[1])),
				fmt.Sprintf(`%s.*?href=['"]([^'"]+)['"].*?target="_blank"`, regexp.QuoteMeta(match[1])),
				fmt.Sprintf(`%s.*?<a[^>]+href=['"]([^'"]+)['"]`, regexp.QuoteMeta(match[1])),
			}

			ep := episodeMap[epID]
			for _, pattern := range linkPatterns {
				re := regexp.MustCompile(pattern)
				linkMatch := re.FindStringSubmatch(bodyString)
				if len(linkMatch) > 1 {
					ep.Links = append(ep.Links, linkMatch[1])
					episodeMap[epID] = ep
					break
				}
			}

			if len(episodeMap[epID].Links) > 0 {
				episodeMap[epID] = Episode{ID: epID, Links: episodeMap[epID].Links}
			}
		}

		// Convert map to slice
		var episodes []Episode
		for _, ep := range episodeMap {
			episodes = append(episodes, ep)
		}

		if len(episodes) > 0 {
			info.Seasons[season] = episodes
		}
	}

	return info, nil
}
