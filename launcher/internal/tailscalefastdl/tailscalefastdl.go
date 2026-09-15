// Package tailscalefastdl configures Tailscale Funnel as the public endpoint
// for TF2's downloadable content. Funnel proxies to the launcher's loopback
// HTTP listener, so Tailscale never needs permission to read the game files.
package tailscalefastdl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	httpsPort = "443"
	urlPath   = "/tf"
)

// Result describes the public endpoint Tailscale Funnel prepared.
type Result struct {
	URL     string
	DNSName string
}

// Authorization is the result of checking Funnel from the settings screen.
// ApprovalURL is present when the tailnet owner still needs to enable Funnel.
type Authorization struct {
	Ready       bool
	ApprovalURL string
}

// ApprovalRequiredError means the tailnet owner has not yet authorized this
// node to use Funnel. URL is safe to open in the operator's browser.
type ApprovalRequiredError struct {
	URL string
}

func (e *ApprovalRequiredError) Error() string {
	return "tailscale Funnel needs approval at " + e.URL
}

// OperatorRequiredError means tailscaled has not authorized this OS user to
// change Serve or Funnel configuration. The launcher itself should remain
// unprivileged; Tailscale needs one privileged command to name its operator.
type OperatorRequiredError struct{}

func (e *OperatorRequiredError) Error() string {
	return "tailscale denied Funnel setup for this user; first run: sudo tailscale set --operator=$USER ; then rerun this launcher with -setup-funnel"
}

type status struct {
	BackendState string `json:"BackendState"`
	AuthURL      string `json:"AuthURL"`
	Self         struct {
		DNSName string `json:"DNSName"`
	} `json:"Self"`
}

type runner func(context.Context, string, ...string) ([]byte, error)

var funnelApprovalURL = regexp.MustCompile(`https://login\.tailscale\.com/f/funnel\?[^\s]+`)

// AuthorizeContainer signs the bundled Compose sidecar in and publishes its
// FastDL route. The sidecar deliberately runs only tailscaled: unlike
// containerboot it can remain alive while a person follows an interactive
// login URL, and the named state volume remembers that login afterwards.
func AuthorizeContainer(ctx context.Context, socket, hostname string, localPort int) (Authorization, error) {
	executable, err := executablePath()
	if err != nil {
		return Authorization{}, err
	}
	return authorizeContainer(ctx, executable, socket, hostname, localPort, run)
}

func authorizeContainer(ctx context.Context, executable, socket, hostname string, localPort int, baseCommand runner) (Authorization, error) {
	command := socketCommand(socket, baseCommand)
	state, err := readStatus(ctx, executable, command)
	if err != nil {
		return Authorization{}, err
	}
	if state.BackendState != "Running" {
		if state.AuthURL == "" {
			// A short timeout starts interactive login without holding the UI
			// request open until the person has visited the returned page.
			_, _ = command(ctx, executable, "up", "--json", "--timeout=4s",
				"--hostname="+hostname, "--accept-dns=false")
			state, err = waitForAuthURL(ctx, executable, command)
			if err != nil {
				return Authorization{}, err
			}
		}
		if !strings.HasPrefix(state.AuthURL, "https://login.tailscale.com/") {
			return Authorization{}, fmt.Errorf("tailscale is waiting for login but returned an invalid approval URL")
		}
		return Authorization{ApprovalURL: state.AuthURL}, nil
	}

	configured, err := configure(ctx, executable, localPort, command)
	if err != nil {
		if approval, ok := errors.AsType[*ApprovalRequiredError](err); ok {
			return Authorization{ApprovalURL: approval.URL}, nil
		}
		return Authorization{}, err
	}
	if configured.URL == "" {
		return Authorization{}, errors.New("tailscale Funnel did not return a FastDL URL")
	}
	return Authorization{Ready: true}, nil
}

func socketCommand(socket string, command runner) runner {
	return func(ctx context.Context, executable string, args ...string) ([]byte, error) {
		withSocket := append([]string{"--socket=" + socket}, args...)
		return command(ctx, executable, withSocket...)
	}
}

func readStatus(ctx context.Context, executable string, command runner) (status, error) {
	raw, err := command(ctx, executable, "status", "--json")
	if err != nil {
		return status{}, fmt.Errorf("cannot reach the bundled Tailscale service: %w", err)
	}
	var state status
	if err := json.Unmarshal(raw, &state); err != nil {
		return status{}, fmt.Errorf("cannot read Tailscale status: %w", err)
	}
	return state, nil
}

func waitForAuthURL(ctx context.Context, executable string, command runner) (status, error) {
	for range 20 {
		state, err := readStatus(ctx, executable, command)
		if err != nil {
			return status{}, err
		}
		if state.AuthURL != "" || state.BackendState == "Running" {
			return state, nil
		}
		select {
		case <-ctx.Done():
			return status{}, fmt.Errorf("tailscale did not produce a login page: %w", ctx.Err())
		case <-time.After(150 * time.Millisecond):
		}
	}
	return status{}, errors.New("tailscale is starting sign-in; wait a moment and press Set up / check Funnel again")
}

// Authorize checks whether this tailnet permits Funnel without needing an
// installed game. When permission is already present it briefly publishes a
// fixed setup message and immediately removes it. When permission is absent,
// Tailscale returns the browser URL the owner must visit.
func Authorize(ctx context.Context) (Authorization, error) {
	executable, err := executablePath()
	if err != nil {
		return Authorization{}, err
	}
	return authorize(ctx, executable, run)
}

func authorize(ctx context.Context, executable string, command runner) (Authorization, error) {
	const setupPath = "/tf2ap-funnel-setup"
	_, err := command(ctx, executable, "funnel", "--yes", "--bg", "--https="+httpsPort,
		"--set-path="+setupPath, "text:TF2 Archipelago Funnel setup")
	if err != nil {
		if approvalURL := funnelApprovalURL.FindString(err.Error()); approvalURL != "" {
			return Authorization{ApprovalURL: approvalURL}, nil
		}
		if operatorRequired(err) {
			return Authorization{}, &OperatorRequiredError{}
		}
		return Authorization{}, fmt.Errorf("cannot enable Tailscale Funnel: %w", err)
	}
	if _, err := command(ctx, executable, "funnel", "--https="+httpsPort,
		"--set-path="+setupPath, "off"); err != nil {
		return Authorization{}, fmt.Errorf("funnel is enabled, but the setup check could not be removed: %w", err)
	}
	return Authorization{Ready: true}, nil
}

func operatorRequired(err error) bool {
	detail := strings.ToLower(err.Error())
	return strings.Contains(detail, "serve config denied") ||
		strings.Contains(detail, "tailscale set --operator=")
}

// Configure finds the installed Tailscale client, verifies that it is signed
// in, and persistently publishes the launcher's loopback HTTP listener.
func Configure(ctx context.Context, localPort int) (Result, error) {
	executable, err := executablePath()
	if err != nil {
		return Result{}, err
	}
	return configure(ctx, executable, localPort, run)
}

// Disable removes only the FastDL mount. Other Funnel and Serve paths on the
// same node belong to their own applications and remain configured.
func Disable(ctx context.Context) error {
	executable, err := executablePath()
	if err != nil {
		return err
	}
	return disable(ctx, executable, run)
}

func disable(ctx context.Context, executable string, command runner) error {
	if _, err := command(ctx, executable, "funnel", "--https="+httpsPort,
		"--set-path="+urlPath, "off"); err != nil {
		return fmt.Errorf("cannot remove the FastDL Funnel route: %w", err)
	}
	return nil
}

func configure(ctx context.Context, executable string, localPort int, command runner) (Result, error) {
	raw, err := command(ctx, executable, "status", "--json")
	if err != nil {
		return Result{}, fmt.Errorf("cannot read Tailscale status: %w", err)
	}
	var state status
	if err := json.Unmarshal(raw, &state); err != nil {
		return Result{}, fmt.Errorf("cannot read Tailscale status: %w", err)
	}
	if state.BackendState != "Running" {
		return Result{}, fmt.Errorf("tailscale is not connected (state %q); start Tailscale and sign in on this machine", state.BackendState)
	}
	dnsName := strings.TrimSuffix(strings.TrimSpace(state.Self.DNSName), ".")
	if dnsName == "" {
		return Result{}, errors.New("tailscale has no MagicDNS name; enable MagicDNS for the tailnet")
	}

	// Funnel removes the mount point before proxying. Give the proxy target the
	// same path so /tf/maps/x.bsp reaches the local handler as /tf/maps/x.bsp.
	target := "http://" + net.JoinHostPort("127.0.0.1", strconv.Itoa(localPort)) + urlPath
	if _, err := command(ctx, executable, "funnel", "--yes", "--bg", "--https="+httpsPort,
		"--set-path="+urlPath, target); err != nil {
		if approvalURL := funnelApprovalURL.FindString(err.Error()); approvalURL != "" {
			return Result{}, &ApprovalRequiredError{URL: approvalURL}
		}
		return Result{}, fmt.Errorf("cannot publish the download server: %w", err)
	}

	return Result{
		URL:     "https://" + dnsName + urlPath,
		DNSName: dnsName,
	}, nil
}

func executablePath() (string, error) {
	if path, err := exec.LookPath("tailscale"); err == nil {
		return path, nil
	}
	if path := installedExecutablePath(); path != "" {
		return path, nil
	}
	return "", errors.New("tailscale is not installed; install it and sign in on this machine")
}

func run(ctx context.Context, executable string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, executable, args...).CombinedOutput()
	if err == nil {
		return output, nil
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		detail = err.Error()
	}
	return nil, errors.New(detail)
}
