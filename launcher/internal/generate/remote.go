package generate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const remoteTimeout = 3*time.Minute + 15*time.Second

// RemoteResult is one archive returned by the Compose-network generator.
// The caller owns Body and closes it after sending the download to the browser.
type RemoteResult struct {
	Name string
	Body io.ReadCloser
}

// Remote asks the pinned Archipelago container to generate from one player
// file. The service address is runtime wiring supplied by Compose, never a
// player setting.
func Remote(ctx context.Context, serviceURL string, playerFile []byte) (RemoteResult, error) {
	if strings.TrimSpace(serviceURL) == "" {
		return RemoteResult{}, fmt.Errorf("the Docker seed generator is not configured")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(serviceURL, "/")+"/generate", bytes.NewReader(playerFile))
	if err != nil {
		return RemoteResult{}, fmt.Errorf("cannot prepare seed generation: %w", err)
	}
	request.Header.Set("Content-Type", "application/yaml")
	client := &http.Client{Timeout: remoteTimeout}
	response, err := client.Do(request)
	if err != nil {
		return RemoteResult{}, fmt.Errorf("the bundled seed generator did not answer: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		defer func() { _ = response.Body.Close() }()
		detail, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return RemoteResult{}, fmt.Errorf("the bundled seed generator returned %s: %s",
			response.Status, strings.TrimSpace(string(detail)))
	}

	name := "tf2-seed.zip"
	if _, params, parseErr := mime.ParseMediaType(response.Header.Get("Content-Disposition")); parseErr == nil {
		if candidate := filepath.Base(params["filename"]); candidate != "." && strings.HasSuffix(candidate, ".zip") {
			name = candidate
		}
	}
	return RemoteResult{Name: name, Body: response.Body}, nil
}
