package webapi

import (
	"context"
	"time"

	"connectrpc.com/connect"

	launcherv1 "github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1"
	"github.com/m-this/tf2-archipelago/launcher/internal/tailscalefastdl"
)

// The three services are the wire and nothing else: each procedure is one call
// into App and one translation of what it answered. There is no state here, and
// no decision that is not already made in the package beside them.

// funnelGrace bounds the wait for Tailscale to answer about this tailnet. The
// player is looking at a button while it runs.
const funnelGrace = 4 * time.Second

// LauncherRPC is LauncherService over one App.
type LauncherRPC struct{ App *App }

// Local address discovery owns its own timeout and does no work on behalf of
// the browser's request.
//
//nolint:contextcheck // Address discovery owns its timeout instead of the request.
func (s LauncherRPC) GetSnapshot(context.Context, *connect.Request[launcherv1.GetSnapshotRequest]) (*connect.Response[launcherv1.GetSnapshotResponse], error) {
	return connect.NewResponse(&launcherv1.GetSnapshotResponse{
		Snapshot: s.App.Snapshot().Proto(),
	}), nil
}

func (s LauncherRPC) Start(context.Context, *connect.Request[launcherv1.StartRequest]) (*connect.Response[launcherv1.StartResponse], error) {
	s.App.Start()
	return connect.NewResponse(&launcherv1.StartResponse{}), nil
}

// Stop waits on two processes going away, which is longer than a button press
// should hold a request open. The stream says when it happened.
func (s LauncherRPC) Stop(context.Context, *connect.Request[launcherv1.StopRequest]) (*connect.Response[launcherv1.StopResponse], error) {
	go s.App.Stop()
	return connect.NewResponse(&launcherv1.StopResponse{}), nil
}

func (s LauncherRPC) Restart(context.Context, *connect.Request[launcherv1.RestartRequest]) (*connect.Response[launcherv1.RestartResponse], error) {
	s.App.Restart()
	return connect.NewResponse(&launcherv1.RestartResponse{}), nil
}

func (s LauncherRPC) Quit(context.Context, *connect.Request[launcherv1.QuitRequest]) (*connect.Response[launcherv1.QuitResponse], error) {
	s.App.Quit()
	return connect.NewResponse(&launcherv1.QuitResponse{}), nil
}

//nolint:contextcheck // Dialling RCON owns its timeout instead of the request.
func (s LauncherRPC) SendRcon(_ context.Context, request *connect.Request[launcherv1.SendRconRequest]) (*connect.Response[launcherv1.SendRconResponse], error) {
	s.App.SendRCON(request.Msg.GetCommand())
	return connect.NewResponse(&launcherv1.SendRconResponse{}), nil
}

// SetMission is the same command the console takes, so the plugin has one way
// in and the browser is not a second.
//
//nolint:contextcheck // Dialling RCON owns its timeout instead of the request.
func (s LauncherRPC) SetMission(_ context.Context, request *connect.Request[launcherv1.SetMissionRequest]) (*connect.Response[launcherv1.SetMissionResponse], error) {
	s.App.SendRCON("sm_ap_mission " + request.Msg.GetPopFile())
	return connect.NewResponse(&launcherv1.SetMissionResponse{}), nil
}

// ApproveFunnel answers with the page the player has to visit, or the word
// saying the tailnet already allows Funnel and there is nothing to do.
func (s LauncherRPC) ApproveFunnel(ctx context.Context, _ *connect.Request[launcherv1.ApproveFunnelRequest]) (*connect.Response[launcherv1.ApproveFunnelResponse], error) {
	ctx, cancel := context.WithTimeout(ctx, funnelGrace)
	defer cancel()
	result, err := tailscalefastdl.Authorize(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	response := &launcherv1.ApproveFunnelResponse{ApprovalUrl: result.ApprovalURL}
	if result.ApprovalURL == "" {
		response.Message = "Tailscale Funnel is ready for this tailnet."
	}
	return connect.NewResponse(response), nil
}

// SettingsRPC is SettingsService over one App.
type SettingsRPC struct{ App *App }

func (s SettingsRPC) OpenSettings(_ context.Context, request *connect.Request[launcherv1.OpenSettingsRequest]) (*connect.Response[launcherv1.OpenSettingsResponse], error) {
	s.App.OpenSettings(request.Msg.GetPage())
	return connect.NewResponse(&launcherv1.OpenSettingsResponse{Screen: s.App.Screen().Proto()}), nil
}

// ChangeSetting refuses a value form cannot apply, with the words form chose:
// the player typed it, so the message is about what they typed.
func (s SettingsRPC) ChangeSetting(_ context.Context, request *connect.Request[launcherv1.ChangeSettingRequest]) (*connect.Response[launcherv1.ChangeSettingResponse], error) {
	if err := s.App.Change(changeFrom(request.Msg.GetChange())); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&launcherv1.ChangeSettingResponse{Screen: s.App.Screen().Proto()}), nil
}

// Actions start downloads, repairs and server lifecycle work that deliberately
// survives the request that asked for them.
//
//nolint:contextcheck // A dispatched action can outlive the request.
func (s SettingsRPC) DispatchAction(_ context.Context, request *connect.Request[launcherv1.DispatchActionRequest]) (*connect.Response[launcherv1.DispatchActionResponse], error) {
	if err := s.App.Dispatch(request.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	return connect.NewResponse(&launcherv1.DispatchActionResponse{Screen: s.App.Screen().Proto()}), nil
}

// SaveSettings reports a refusal in the response rather than as an error.
// Settings nothing could act on are for the player to fix with their answers
// still in front of them, which is why the screen comes back with it.
//
//nolint:contextcheck // Post-save lifecycle work outlives the request.
func (s SettingsRPC) SaveSettings(_ context.Context, request *connect.Request[launcherv1.SaveSettingsRequest]) (*connect.Response[launcherv1.SaveSettingsResponse], error) {
	response := &launcherv1.SaveSettingsResponse{Saved: true}
	if err := s.App.SaveSettings(request.Msg.GetRestart()); err != nil {
		response.Saved, response.Refusal = false, err.Error()
	}
	response.Screen = s.App.Screen().Proto()
	return connect.NewResponse(response), nil
}

func (s SettingsRPC) CancelSettings(context.Context, *connect.Request[launcherv1.CancelSettingsRequest]) (*connect.Response[launcherv1.CancelSettingsResponse], error) {
	s.App.CancelSettings()
	return connect.NewResponse(&launcherv1.CancelSettingsResponse{}), nil
}
