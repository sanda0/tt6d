package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Writer wraps an io.Writer and tracks progress
type Writer struct {
	writer     io.Writer
	total      int64
	written    int64
	filename   string
	index      int
	totalFiles int
	lastUpdate time.Time
}

// New creates a new progress writer
func New(writer io.Writer, total int64, filename string, index, totalFiles int) *Writer {
	return &Writer{
		writer:     writer,
		total:      total,
		filename:   filename,
		index:      index,
		totalFiles: totalFiles,
		lastUpdate: time.Now(),
	}
}

func (pw *Writer) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.written += int64(n)

	// Update progress bar every 100ms to avoid too frequent updates
	if time.Since(pw.lastUpdate) >= 100*time.Millisecond || pw.written == pw.total {
		pw.displayProgress()
		pw.lastUpdate = time.Now()
	}

	return n, err
}

var (
	mutex sync.Mutex
)

func (pw *Writer) displayProgress() {
	percentage := float64(pw.written) / float64(pw.total) * 100
	if pw.total == 0 {
		percentage = 0
	}

	// Create progress bar
	barWidth := 30
	filled := int(percentage * float64(barWidth) / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	// Format file size
	writtenMB := float64(pw.written) / (1024 * 1024)
	totalMB := float64(pw.total) / (1024 * 1024)

	// Lock to prevent progress bars from mixing
	mutex.Lock()
	defer mutex.Unlock()

	// Move cursor to the correct line based on the worker index
	fmt.Printf("\033[%d;0H\033[K[%d/%d] %s [%s] %.1f%% (%.1f/%.1f MB)",
		pw.index, pw.index, pw.totalFiles, pw.filename, bar, percentage, writtenMB, totalMB)

	// If download is complete, mark with a checkmark and clear the line
	if pw.written == pw.total {
		fmt.Printf(" ✓")
		// Clear this progress bar after a short delay
		go func() {
			time.Sleep(2 * time.Second)
			mutex.Lock()
			fmt.Printf("\033[%d;0H\033[K", pw.index) // Clear the line
			mutex.Unlock()
		}()

		// Move cursor back to bottom
		fmt.Printf("\033[%d;0H", pw.totalFiles+1)
	}
}
