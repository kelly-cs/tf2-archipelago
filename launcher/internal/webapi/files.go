package webapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"

	"github.com/m-this/tf2-archipelago/launcher/internal/assets"
	"github.com/m-this/tf2-archipelago/launcher/internal/debugbundle"
	launcherv1 "github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1"
	"github.com/m-this/tf2-archipelago/launcher/internal/generate"
	"github.com/m-this/tf2-archipelago/launcher/internal/settings"
	"github.com/m-this/tf2-archipelago/launcher/internal/winproc"
)

// bundleChunk is how much of the zip travels in one message. The bundle is a
// few megabytes and the connection is loopback, so this is only about keeping
// one message under Connect's default ceiling.
const bundleChunk = 64 << 10

// FilePath answers where the player's copy of something is, making it first
// where making it is what the button means: the player file and the seed are
// written from the settings on the screen, not read from disk.
func (a *App) FilePath(ctx context.Context, target launcherv1.FileTarget) (string, error) {
	if target == launcherv1.FileTarget_FILE_TARGET_SETTINGS_FILE {
		return settings.Path()
	}
	s, err := a.DraftSettings()
	if err != nil {
		return "", err
	}
	switch target {
	case launcherv1.FileTarget_FILE_TARGET_PLAYER_FILE:
		return settings.WritePlayerFile(s, assets.ArchipelagoVersion)
	case launcherv1.FileTarget_FILE_TARGET_INSTALL_ROOT:
		return s.InstallRoot, nil
	case launcherv1.FileTarget_FILE_TARGET_GENERATED_SEED:
		result, err := generate.Run(ctx, generate.Options{
			Settings: s, AppDir: s.ArchipelagoDir,
			Apworld: assets.Apworld(), ArchipelagoVersion: assets.ArchipelagoVersion,
		})
		if err != nil {
			return "", err
		}
		return result.Archive, nil
	case launcherv1.FileTarget_FILE_TARGET_UNSPECIFIED, launcherv1.FileTarget_FILE_TARGET_SETTINGS_FILE:
	}
	return "", fmt.Errorf("no file target %d", target)
}

// FilesRPC is FilesService over one App.
type FilesRPC struct {
	App *App

	// open shows a path to the player. A field rather than a call so a test can
	// ask what was opened without a desktop; nil means winproc.Open.
	open func(string) error
}

// NewFilesRPC wires the service to the desktop's own handler.
func NewFilesRPC(app *App) FilesRPC { return FilesRPC{App: app, open: winproc.Open} }

// ShowFile hands a path to the desktop, because a browser cannot open a folder.
// It answers with what it opened either way, so a machine with no handler for a
// .yaml still tells the player where to look.
func (s FilesRPC) ShowFile(ctx context.Context, request *connect.Request[launcherv1.ShowFileRequest]) (*connect.Response[launcherv1.ShowFileResponse], error) {
	path, err := s.App.FilePath(ctx, request.Msg.GetTarget())
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	}
	if err := s.open(path); err != nil {
		s.App.Say("could not open %s: %v", path, err)
	}
	return connect.NewResponse(&launcherv1.ShowFileResponse{Path: path}), nil
}

// DownloadDebugBundle streams the zip. The first message names it and carries
// no bytes; every message after it carries bytes and no name.
//
//nolint:contextcheck // Bundle collection owns its bridge timeout.
func (s FilesRPC) DownloadDebugBundle(_ context.Context, _ *connect.Request[launcherv1.DownloadDebugBundleRequest], stream *connect.ServerStream[launcherv1.DownloadDebugBundleResponse]) error {
	settingsNow, err := s.App.DraftSettings()
	if err != nil {
		return connect.NewError(connect.CodeFailedPrecondition, err)
	}
	path, err := debugbundle.Write(settingsNow, assets.Versions(), time.Now())
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	file, err := os.Open(path) //nolint:gosec // The path is the launcher's own bundle, not the browser's.
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	defer func() { _ = file.Close() }()

	if err := stream.Send(&launcherv1.DownloadDebugBundleResponse{Filename: filepath.Base(path)}); err != nil {
		return err
	}
	buffer := make([]byte, bundleChunk)
	for {
		read, err := file.Read(buffer)
		if read > 0 {
			if err := stream.Send(&launcherv1.DownloadDebugBundleResponse{Chunk: buffer[:read]}); err != nil {
				return err
			}
		}
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
	}
}
