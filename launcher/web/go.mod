// This file is a fence, not a module anybody builds.
//
// npm packages ship Go source: flatted, an eslint dependency, has a whole
// package under node_modules. Without a go.mod here, `./...` from the
// repository root builds, vets, tests and vulnerability-scans it as ours, and
// `make lint` fails on somebody else's code. A directory holding a go.mod is
// not part of its parent module, which is all this needs to do.
//
// No go directive on purpose: go.mod at the root owns the Go version, and a
// second one here would be a second place to say it that nothing checks.
module github.com/m-this/tf2-archipelago/launcher/web
