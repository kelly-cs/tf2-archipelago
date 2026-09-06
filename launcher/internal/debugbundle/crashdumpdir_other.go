//go:build !windows

package debugbundle

func systemCrashDumpDir() string {
	return ""
}
