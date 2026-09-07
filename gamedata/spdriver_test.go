package gamedata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m-this/tf2-mvm-bots-go/spshell"
)

/*
Running a plugin function instead of reading it.

The tests beside this one check the plugin by looking for substrings in its
source: that WeaponBuffs_AttackEnemyProjectiles still contains "chosen =
projectile", that the sweep still mentions TE_SetupSparks. They were written
because nothing in SourcePawn can check itself, and they are the best that can
be done by reading. It is not much. Renaming a local breaks them without
changing behaviour, and changing the arithmetic passes them as long as the names
survive, which is the wrong way round for both.

So the toolchain the defender mod built for its generated code is pointed at
this plugin's hand-written code. spcomp compiles a driver and SourcePawn's
standalone VM runs it, and what comes back is what the function computes, on
inputs this test chose. No game server, no map, no client.

# Why the function is pasted rather than included

weapon_buffs.inc cannot be compiled on its own. It opens with sourcemod,
sdkhooks, tf2 and ripext, and the standalone VM has none of them. So the driver
takes the text of one function, by name, with the #defines it reads, and
compiles that.

Locating by name is the part these tests still share with the ones that read
source, and it is the part that is fine: a function that has been renamed fails
loudly here rather than silently passing. What is no longer shared is the
assertion. This one runs the arithmetic.

The clean end of this is a weapon_buffs_math.inc that includes nothing and is
included by both the plugin and the driver, and it is bead apw-form. Until then
this proves the harness against the code as it stands.
*/

// requireEnv turns an absent toolchain from a skip into a failure. make check
// sets it; a developer with no clang gets the skip and a message naming what to
// run. The defender mod has its own variable for its own gate, which is why
// spshell takes the name rather than owning it.
const requireEnv = "TF2AP_REQUIRE_SPSHELL"

// buffsSource is the plugin file every driver here takes its functions from.
const buffsSource = "../plugin/scripting/tf2_archipelago/weapon_buffs.inc"

/*
	driver is a standalone plugin built out of pieces of the real one

defines are the #define lines lifted from the top of weapon_buffs.inc, so a
constant is never transcribed here: changing ProjectileDestructionBaseCooldown
in the plugin changes what this runs.

funcs are the functions under test, by signature, pasted whole.

main is the body that calls them and prints a cell per answer.
*/
type driver struct {
	defines []string
	funcs   []string
	main    string
}

func (d driver) source(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("#pragma semicolon 1\n#pragma newdecls required\n\n")
	// spshell's own builtin, and the only one any driver here needs. A float
	// goes out as view_as<int> so the bits arrive rather than a rounded decimal.
	b.WriteString("native void printnum(int n);\n\n")
	for _, name := range d.defines {
		b.WriteString(defineFrom(t, buffsSource, name))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	for _, signature := range d.funcs {
		b.WriteString(sourceFunctionWithSignature(t, buffsSource, signature))
		b.WriteString("\n\n")
	}
	b.WriteString("public int main()\n{\n")
	b.WriteString(d.main)
	b.WriteString("\n    return 0;\n}\n")
	return b.String()
}

// run compiles the driver and returns the cells it printed, in order.
func (d driver) run(t *testing.T) []int32 {
	t.Helper()
	tc := spshell.ForTestRequiring(t, requireEnv)

	path := filepath.Join(t.TempDir(), "driver.sp")
	if err := os.WriteFile(path, []byte(d.source(t)), 0o600); err != nil {
		t.Fatal(err)
	}
	cells, err := tc.Run(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("running %s: %v\n\n%s", path, err, d.source(t))
	}
	return cells
}

/*
	The projectile destruction cooldown is the arithmetic, not the wording

One level destroys a projectile every ProjectileDestructionBaseCooldown seconds,
each level after takes a step off, and it never goes below the floor. Three
constants and a max, and every one of them is a number a reader can get wrong
while leaving the source looking right.

The expected values are worked out here from the same #defines the plugin
carries, so this does not pin today's numbers: it pins the rule. Changing the
base cooldown moves both sides. Changing the subtraction to a division moves one.
*/
func TestProjectileDestructionCooldownIsTheDeclaredCurve(t *testing.T) {
	base := floatDefine(t, buffsSource, "ProjectileDestructionBaseCooldown")
	step := floatDefine(t, buffsSource, "ProjectileDestructionCooldownStep")
	floor := floatDefine(t, buffsSource, "ProjectileDestructionMinimumCooldown")

	levels := []int{1, 2, 3, 4, 5, 8, 20}
	var calls strings.Builder
	for _, level := range levels {
		fmt.Fprintf(&calls, "    printnum(view_as<int>(WeaponBuffs_ProjectileDestructionCooldown(%d)));\n", level)
	}

	got := driver{
		defines: []string{
			"ProjectileDestructionBaseCooldown",
			"ProjectileDestructionCooldownStep",
			"ProjectileDestructionMinimumCooldown",
		},
		funcs: []string{"static float WeaponBuffs_ProjectileDestructionCooldown"},
		main:  calls.String(),
	}.run(t)

	if len(got) != len(levels) {
		t.Fatalf("%d levels went in and %d answers came out", len(levels), len(got))
	}
	for i, level := range levels {
		want := base - float32(level-1)*step
		if want < floor {
			want = floor
		}
		if bitsToFloat(got[i]) != want {
			t.Errorf("level %d cools down in %v, wanted %v", level, bitsToFloat(got[i]), want)
		}
	}
}

/*
	A passive effect is the three the plugin names and nothing else

WeaponBuffs_IsPassiveEffect decides which effects are applied at spawn rather
than on a hit or a kill, and getting it wrong is silent: an effect that falls out
of the passive set simply never applies, and the player reports a buff that does
nothing.

Every effect ID the generated table holds is tried, not the three that are
expected, because what matters is the answer for the ones nobody thought about.
*/
func TestOnlyTheNamedEffectsArePassive(t *testing.T) {
	passive := map[int]bool{
		intDefine(t, buffsSource, "MoveSpeedEffect"):         true,
		intDefine(t, buffsSource, "JumpHeightEffect"):        true,
		intDefine(t, buffsSource, "ActiveHealthRegenEffect"): true,
	}

	count := intDefine(t, "../plugin/scripting/tf2_archipelago/weapon_buffs_data.inc", "WeaponEffectCount")
	var calls strings.Builder
	for effect := range count {
		fmt.Fprintf(&calls, "    printnum(WeaponBuffs_IsPassiveEffect(%d) ? 1 : 0);\n", effect)
	}

	got := driver{
		defines: []string{"MoveSpeedEffect", "JumpHeightEffect", "ActiveHealthRegenEffect"},
		funcs:   []string{"static bool WeaponBuffs_IsPassiveEffect"},
		main:    calls.String(),
	}.run(t)

	if len(got) != count {
		t.Fatalf("%d effects went in and %d answers came out", count, len(got))
	}
	for effect := range count {
		if want := passive[effect]; (got[effect] == 1) != want {
			t.Errorf("effect %d is passive=%v, wanted %v", effect, got[effect] == 1, want)
		}
	}
}
