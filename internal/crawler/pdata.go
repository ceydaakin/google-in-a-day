package crawler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ceydaakin/google-in-a-day/internal/index"
)

// SavePData writes the inverted index to a flat file in the format:
// word url origin depth frequency
// One entry per word-document pair, sorted by word.
func SavePData(path string, docs []index.Document) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	var lines []string
	for _, doc := range docs {
		for word, freq := range doc.WordFreq {
			line := fmt.Sprintf("%s %s %s %d %d", word, doc.URL, doc.OriginURL, doc.Depth, freq)
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
