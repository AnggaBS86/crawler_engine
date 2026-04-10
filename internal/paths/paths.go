package paths

import (
	"errors"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var invalidFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func DefaultSnapshotFilename(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "snapshot-" + time.Now().UTC().Format("20060102T150405Z") + ".html"
	}

	host := strings.ToLower(u.Hostname())
	host = invalidFilenameChars.ReplaceAllString(host, "_")
	if host == "" {
		host = "snapshot"
	}

	return host + "-" + time.Now().UTC().Format("20060102T150405Z") + ".html"
}

func SafeJoin(baseDir, filename string) (string, error) {
	if strings.TrimSpace(filename) == "" {
		return "", errors.New("empty filename")
	}

	clean := filepath.Clean(filename)
	if filepath.IsAbs(clean) {
		return "", errors.New("absolute paths not allowed")
	}
	if strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", errors.New("path traversal not allowed")
	}

	return filepath.Join(baseDir, clean), nil
}

func Dir(p string) string {
	return filepath.Dir(p)
}
