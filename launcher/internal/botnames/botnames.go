/*
Package botnames is the pool the defender bots draw their names from.

The mod reads configs/defenderbots/bot_names.txt, one name per line, at every
map start, and gives each bot it seats a name drawn from it. bot_names.txt here
is that file: embedded so the launcher can offer the list, and copied into the
mod's package by deploy/bots/build.sh so the server has it whether or not a
launcher ever wrote one.

What a player changes is kept as the difference from that file rather than as a
copy of it: the names taken out, and the names added. A list added to in a later
release then reaches everybody, instead of being frozen the day somebody first
opened the page.
*/
package botnames

import (
	_ "embed"
	"slices"
	"strings"
)

//go:embed bot_names.txt
var shippedFile string

/*
NameMax is the longest name that survives the trip.

The game truncates a player name at 32 bytes including the terminator, and a
name cut in half reads as a bug in the launcher rather than as a limit of the
game. Refused at the door instead.
*/
const NameMax = 31

// AddedMax bounds the list somebody can type. The draw is one name per bot and
// a server holds 32 players, so a hundred names is already more than a run can
// show; this is the point past which the page has stopped being readable.
const AddedMax = 64

// Shipped is the pool as it comes, in file order.
func Shipped() []string {
	return parse(shippedFile)
}

/*
Pool is what the bots actually draw from: what was shipped, less what was taken
out, plus what was added.

Order is the shipped order then the added ones, so a player reading the file
finds their own names at the end rather than sorted into somebody else's list.
A name that is both shipped and added appears once: the mod draws by index, and
a duplicate is a name that comes up twice as often for no stated reason.
*/
func Pool(excluded, added []string) []string {
	out := make([]string, 0, len(shippedFile)/8+len(added))
	for _, name := range Shipped() {
		if !slices.Contains(excluded, name) {
			out = append(out, name)
		}
	}
	for _, name := range added {
		name = strings.TrimSpace(name)
		if name != "" && !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	return out
}

// Render is the file the mod reads. An empty pool is written as an empty file
// rather than skipped: the mod says "You forgot to give me a name!" on every
// bot, which is a player's answer showing up in the game, not a failure.
func Render(excluded, added []string) string {
	var b strings.Builder
	b.WriteString("// Managed by tf2ap. Edits here are replaced the next time the launcher starts.\n")
	for _, name := range Pool(excluded, added) {
		b.WriteString(name)
		b.WriteString("\n")
	}
	return b.String()
}

func parse(body string) []string {
	var out []string
	for line := range strings.SplitSeq(body, "\n") {
		if name := strings.TrimSpace(line); name != "" && !strings.HasPrefix(name, "//") {
			out = append(out, name)
		}
	}
	return out
}
