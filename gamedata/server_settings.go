package gamedata

/*
A server setting is a lever on the whole server that arrives as an item.

The health scale and the bot lineup are settings a host types before the run.
These are the same kind of lever handed to the multiworld instead: somebody
else opens a chest and the rule of the game changes for everybody on the slot.

State, not an effect. Granting one twice is granting it once, so it goes in the
unlock set and is resent freely, which is what makes it survive a map change
and a reconnect.
*/

// ServerSettingID identifies one setting. It alone decides the item id, so a
// value here never changes.
type ServerSettingID uint16

const (
	// ServerSettingGrapplingHook gives every player the Mannpower hook. One
	// convar, granted once, and the largest change to how a map plays of the
	// four Peppy asked for.
	ServerSettingGrapplingHook ServerSettingID = 1
)

/*
ServerSetting is one lever. Key is the word on the wire to the plugin, which
matches on it and never sees the id, and the plugin owns which convar it is:
the wire says what was granted, not how to apply it.
*/
type ServerSetting struct {
	ID   ServerSettingID
	Key  string
	Name string
}

// ServerSettings is every lever the multiworld can hand over. The other three
// in Peppy's post are progressives and are not built.
var ServerSettings = []ServerSetting{
	{ServerSettingGrapplingHook, "grappling_hook", "Grappling Hook"},
}

// ItemName is what the multiworld calls the item that grants this setting.
func (s ServerSetting) ItemName() string { return s.Name }

// ServerSettingByID resolves the payload of an item into the setting the
// plugin is told about.
func ServerSettingByID(id ServerSettingID) (ServerSetting, bool) {
	for _, setting := range ServerSettings {
		if setting.ID == id {
			return setting, true
		}
	}
	return ServerSetting{}, false
}
