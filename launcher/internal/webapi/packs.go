package webapi

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/m-this/tf2-archipelago/gamedata"
	"github.com/m-this/tf2-archipelago/launcher/internal/installer"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
)

func availableCommunityPackNames(folder string) []string {
	paths := installer.AvailableCommunityArchives(settings.KnownCommunityArchives(strings.TrimSpace(folder)))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, filepath.Base(path))
	}
	return result
}

func importedCommunityPackNames(folder string, available []string) []string {
	var imported []string
	for _, name := range available {
		if _, err := os.Stat(filepath.Join(folder, name+".imported")); err == nil {
			imported = append(imported, name)
		}
	}
	return imported
}

func recognizedCommunityArchive(name string) (string, bool) {
	name = strings.ToLower(filepath.Base(name))
	switch name {
	case settings.CommunityPackPotato, settings.CommunityPackMoonlight:
		return name, true
	default:
		return "", false
	}
}

func importCommunityArchive(folder, name string, source io.Reader) (string, error) {
	originalName := name
	name, recognized := recognizedCommunityArchive(name)
	if !recognized {
		return "", fmt.Errorf("unrecognized asset pack %q; choose %s or %s", originalName, settings.CommunityPackPotato, settings.CommunityPackMoonlight)
	}
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", fmt.Errorf("cannot create the asset cache: %w", err)
	}
	temporary, err := os.CreateTemp(folder, ".tf2ap-import-*.zip")
	if err != nil {
		return "", fmt.Errorf("cannot stage %s: %w", name, err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if _, err := io.Copy(temporary, source); err != nil {
		_ = temporary.Close()
		return "", fmt.Errorf("cannot import %s: %w", name, err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("cannot finish %s: %w", name, err)
	}
	if err := installer.ValidateCommunityArchives([]string{temporaryName}, func(string, ...any) {}); err != nil {
		return "", err
	}
	target := filepath.Join(folder, name)
	if err := os.Rename(temporaryName, target); err != nil {
		if removeErr := os.Remove(target); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return "", fmt.Errorf("cannot replace %s: %w", name, removeErr)
		}
		if err := os.Rename(temporaryName, target); err != nil {
			return "", fmt.Errorf("cannot keep %s: %w", name, err)
		}
	}
	if err := os.WriteFile(target+".imported", nil, 0o644); err != nil {
		return "", fmt.Errorf("cannot mark %s as imported: %w", name, err)
	}
	return name, nil
}

// ImportArchive writes one offered asset zip into the folder the settings name
// and answers with the pack it turned out to be. It does not touch the draft:
// AcceptImportedPacks is what does, once every archive has landed.
func (a *App) ImportArchive(name string, source io.Reader) (string, error) {
	s, err := a.DraftSettings()
	if err != nil {
		return "", err
	}
	return importCommunityArchive(s.CommunityContentDir, name, source)
}

// AcceptImportedPacks turns landed archives into settings: community missions
// go on, each pack is added, and every mission it brought comes out of the
// excluded list, because a player who imported a pack meant to play it.
func (a *App) AcceptImportedPacks(names []string) error {
	folder, err := a.DraftSettings()
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.draft == nil {
		return errors.New("settings were closed while the assets were importing")
	}
	a.community = availableCommunityPackNames(folder.CommunityContentDir)
	a.imported = importedCommunityPackNames(folder.CommunityContentDir, a.community)
	a.draft.Settings.MvmCommunityMissions = true
	for _, name := range names {
		if !slices.Contains(a.draft.Settings.CommunityPacks, name) {
			a.draft.Settings.CommunityPacks = append(a.draft.Settings.CommunityPacks, name)
		}
		for _, mission := range gamedata.PlayableMissions() {
			if gamedata.MissionPack(mission.ID) == name {
				a.draft.Settings.MvmExcludedMissions = slices.DeleteFunc(
					a.draft.Settings.MvmExcludedMissions,
					func(popFile string) bool { return popFile == mission.PopFile })
			}
		}
	}
	a.publishLocked(Event{Name: "state", Data: struct{}{}})
	return nil
}
