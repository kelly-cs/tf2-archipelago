// Package release asks GitHub whether a newer launcher has been published, so
// a player running an old build hears about it from the log rather than from
// a bug that was already fixed.
package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// LatestAPI is where GitHub says which release is current.
	LatestAPI = "https://api.github.com/repos/m-this/tf2-archipelago/releases/latest"
	// DownloadPage is where the player goes to get it.
	DownloadPage = "https://github.com/m-this/tf2-archipelago/releases/latest"
	// timeout bounds the whole check: a launcher must start whether or not
	// GitHub answers.
	timeout = 5 * time.Second
)

// Check is one look at the latest release.
type Check struct {
	// Current is this build's version, as the release tagged it. Empty is a
	// hand build, which has nothing to compare against and asks nothing.
	Current string
	// API is the endpoint, replaceable by a test.
	API    string
	Client *http.Client
}

// Latest is the tag of the newest release.
func (c Check) Latest(ctx context.Context) (string, error) {
	api := c.API
	if api == "" {
		api = LatestAPI
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github answered %s", resp.Status)
	}
	var body struct {
		Tag string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Tag == "" {
		return "", fmt.Errorf("the latest release has no tag")
	}
	return body.Tag, nil
}

// Notice is the line worth saying, or empty: nothing for a hand build, for
// a launcher that is current, or when GitHub could not be asked.
func (c Check) Notice(ctx context.Context) string {
	if c.Current == "" {
		return ""
	}
	latest, err := c.Latest(ctx)
	if err != nil || !Newer(latest, c.Current) {
		return ""
	}
	return fmt.Sprintf("tf2ap %s is out and this is %s. Download it from %s", latest, c.Current, DownloadPage)
}

// Newer says whether latest is a later version than current. Both are read as
// vMAJOR.MINOR.PATCH with anything after a dash ignored; a tag that is not a
// version, such as nightly, is never newer.
func Newer(latest, current string) bool {
	a, ok := parse(latest)
	if !ok {
		return false
	}
	b, ok := parse(current)
	if !ok {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func parse(tag string) ([3]int, bool) {
	var out [3]int
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	tag, _, _ = strings.Cut(tag, "-")
	parts := strings.Split(tag, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
