package installer

import (
	"archive/zip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// These are hashes of the complete, full-with-maps ZIPs, not their contents
// after extraction. Update them only with a reviewed community catalog.
// docs/en/community-archives.md records the source, size, and snapshot date.
var communityArchiveSHA256 = map[string]string{
	"archive-assets.zip":   "e7e54f3167b97341d11cf1a1b30f437bf0651fec40e4e1d25232b883cf44bb69",
	"mlarchive-assets.zip": "c6ba6c85c4466f012094388a59e15e4973c532cd1fd3bb2f404f1d7abc980149",
}

const (
	communityMismatchSuffix = ".hash-mismatch"
	communityIgnoreSuffix   = ".ignore-hash-mismatch"
)

// CommunityArchiveHashMismatch names the exact file whose bytes differ from
// the catalog snapshot. A valid ZIP is not enough: its maps and missions may
// have changed since the catalog was reviewed.
type CommunityArchiveHashMismatch struct {
	Name     string
	Expected string
	Actual   string
}

func (e *CommunityArchiveHashMismatch) Error() string {
	return fmt.Sprintf("%s SHA-256 mismatch: expected %s, downloaded %s. This archive may cause missing or unstable missions. Review the hashes, then use Ignore hash mismatch on the Missions page to approve this exact downloaded file", e.Name, e.Expected, e.Actual)
}

func communityArchiveDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func validateCommunityArchive(path string) (bool, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return false, fmt.Errorf("community pack %s is not a valid ZIP: %w", path, err)
	}
	if err := reader.Close(); err != nil {
		return false, fmt.Errorf("cannot close community pack %s: %w", path, err)
	}
	expected, known := communityArchiveSHA256[filepath.Base(path)]
	if !known {
		return false, nil
	}
	actual, err := communityArchiveDigest(path)
	if err != nil {
		return false, fmt.Errorf("cannot hash community pack %s: %w", path, err)
	}
	if actual == expected {
		return false, nil
	}
	approved, err := os.ReadFile(path + communityIgnoreSuffix)
	if err == nil && strings.TrimSpace(string(approved)) == actual {
		return true, nil
	}
	return false, &CommunityArchiveHashMismatch{filepath.Base(path), expected, actual}
}

// HoldMismatchedCommunityArchive moves an already cached or imported ZIP out of
// the usable path. The user can inspect and explicitly approve the held bytes.
func HoldMismatchedCommunityArchive(path string) error {
	_, err := validateCommunityArchive(path)
	var mismatch *CommunityArchiveHashMismatch
	if !errors.As(err, &mismatch) {
		return err
	}
	if err := os.Rename(path, path+communityMismatchSuffix); err != nil {
		return fmt.Errorf("cannot hold mismatched %s: %w", path, err)
	}
	return mismatch
}

// PendingCommunityArchiveHashMismatches names downloaded ZIPs awaiting a
// person's decision. The launcher uses this for the confirmation row.
func PendingCommunityArchiveHashMismatches(archives []string) []string {
	var pending []string
	for _, path := range archives {
		if _, known := communityArchiveSHA256[filepath.Base(path)]; !known {
			continue
		}
		if info, err := os.Stat(path + communityMismatchSuffix); err == nil && info.Mode().IsRegular() {
			pending = append(pending, filepath.Base(path))
		}
	}
	slices.Sort(pending)
	return pending
}

// IgnoreCommunityArchiveHashMismatch is called only by the Missions page's
// Confirm action. Its receipt contains the approved digest, so a later changed
// ZIP cannot inherit the approval.
func IgnoreCommunityArchiveHashMismatch(archives []string) ([]string, error) {
	var approved []string
	for _, path := range archives {
		pending := path + communityMismatchSuffix
		if _, err := os.Stat(pending); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return approved, err
		}
		reader, err := zip.OpenReader(pending)
		if err != nil {
			return approved, fmt.Errorf("cannot approve %s: not a valid ZIP: %w", pending, err)
		}
		if err := reader.Close(); err != nil {
			return approved, err
		}
		digest, err := communityArchiveDigest(pending)
		if err != nil {
			return approved, err
		}
		if err := os.Rename(pending, path); err != nil {
			return approved, fmt.Errorf("cannot use approved %s: %w", pending, err)
		}
		if err := os.WriteFile(path+communityIgnoreSuffix, []byte(digest+"\n"), 0o644); err != nil {
			return approved, fmt.Errorf("cannot record approval for %s: %w", path, err)
		}
		approved = append(approved, filepath.Base(path))
	}
	if len(approved) == 0 {
		return nil, errors.New("no downloaded archive is awaiting hash mismatch approval")
	}
	return approved, nil
}
