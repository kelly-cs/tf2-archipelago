# Patches to the defender bots and their dependencies

`deploy/bots/build.sh` applies these to the pinned upstream checkouts before it
compiles. A patch that no longer applies fails the build. That is the signal
to rebase it on the new upstream, or to drop it because upstream fixed it.

Most defender-mod changes live in the fork named by `DEFENDERBOTS_REPO`, and
`DEFENDERBOTS_VERSION` names its integration branch or release. Small test
fixes may live here temporarily when a launcher build needs to carry them
before the mod has a release; the build applies and compiles them exactly like
dependency patches.

| Patch | Why |
| --- | --- |
| `defenderbots/0001-apply-robot-health-after-template.patch` | Wait for the robot template to settle, scale TF2's authoritative player-resource maximum rather than the raw class field, invalidate TF2's attribute cache, and log any observed mismatch. |
| `tf2attributes/0001-drop-pragma-unused-before-declaration.patch` | `#pragma unused` sits above the function it names, and spcomp 1.12 resolves it before the declaration exists. Nothing compiles without this. |
| `actions/0001-drop-libudis86-and-asm-sources.patch` | Only for a from-source build of the extension. SourceMod removed `public/libudis86` and `public/asm` in commit `e07c120c`; the AMBuilder still lists them. |
| `actions/0002-do-not-treat-sdk-warnings-as-errors.patch` | Only for a from-source build. `-Werror` against the current hl2sdk fails on `#pragma warning` blocks in `vstdlib/random.h`. |
| `actions/0003-terminate-handle-instead-of-assigning-index.patch` | Only for a from-source build. `CBaseHandle::operator=` takes an `IHandleEntity*`, not an index. |

The three Actions patches matter only with `BOTS_BUILD_EXTENSIONS=1`. The
normal build downloads the two extensions from their releases.

Two more facts the build depends on, explained in the scripts. The defender
mod compiles with SourceMod 1.12.0-git7164's `spcomp`, and git7246's segfaults
on it. Actions stays at v3.9.2: v4.0.0 has no TF2 build, and its TF2 source
does not compile.
