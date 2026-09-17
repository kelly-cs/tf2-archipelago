package webapi

import (
	"fmt"
	"regexp"
	"strings"

	apruntime "github.com/m-this/tf2-archipelago/launcher/internal/runtime"
)

var (
	consoleColor = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	qhullWarning = [...]string{
		"qhull precision error: Only 4 facets remain.  Can not merge another",
		"pair.  The convexity constraints may be too strong.  Reduce the",
		"magnitude of 'Cn' or increase the magnitude of 'An'.  For example,",
		"try 'C-0.001' instead of 'C-0.1' or",
	}
)

// Collapse only complete copies of this known multiline warning. Unexpected
// output is preserved, including engine messages appended to its final line.
// The server's console.log and the launcher's disk log retain the raw output.
type logNoise struct {
	pending []apruntime.Line
	seen    bool
	repeats int
}

func (n *logNoise) filter(line apruntime.Line) []apruntime.Line {
	if line.Source != "srcds" {
		return []apruntime.Line{line}
	}
	clean := strings.TrimSpace(consoleColor.ReplaceAllString(line.Text, ""))
	step := len(n.pending)
	matches := clean == qhullWarning[step]
	if step == len(qhullWarning)-1 {
		matches = strings.HasPrefix(clean, qhullWarning[step])
	}
	if matches {
		n.pending = append(n.pending, line)
		if len(n.pending) < len(qhullWarning) {
			return nil
		}
		var out []apruntime.Line
		suppressed := n.seen
		if !n.seen {
			out = append(out, n.pending...)
			n.seen = true
		} else {
			n.repeats++
		}
		n.pending = nil
		tail := strings.TrimSpace(strings.TrimPrefix(clean, qhullWarning[step]))
		if tail != "" {
			out = append(out, n.summary(line)...)
			// The first copy already contains the appended message.
			if suppressed {
				line.Text = tail
				out = append(out, line)
			}
			n.seen = false
		}
		return out
	}
	out := n.summary(line)
	out = append(out, n.pending...)
	n.pending = nil
	n.seen = false
	// A new header after an incomplete block starts a fresh candidate.
	if clean == qhullWarning[0] {
		n.pending = append(n.pending, line)
	} else {
		out = append(out, line)
	}
	return out
}

func (n *logNoise) summary(line apruntime.Line) []apruntime.Line {
	if n.repeats == 0 {
		return nil
	}
	line.Text = fmt.Sprintf("[log] Collapsed %d repeated Qhull warnings; full output is in the server console log.", n.repeats)
	n.repeats = 0
	return []apruntime.Line{line}
}
