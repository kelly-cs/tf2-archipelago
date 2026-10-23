package form

import (
	"strings"
	"testing"

	launcherv1 "github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1"
)

// kindsMax bounds the walk over Kind. Nothing declares how many there are, so
// the test counts them by asking String, and this stops a Kind whose String
// answers for every value from running forever.
const kindsMax = 64

// declaredKinds is every Kind the package declares, found the only way the
// package allows: String names the ones it knows and formats the rest.
func declaredKinds(t *testing.T) []Kind {
	t.Helper()
	kinds := make([]Kind, 0, kindsMax)
	for k := Text; uint8(k) <= kindsMax; k++ {
		if strings.HasPrefix(k.String(), "kind(") {
			return kinds
		}
		kinds = append(kinds, k)
	}
	t.Fatalf("more than %d kinds: String names every value, so nothing can count them", kindsMax)
	return nil
}

// The browser renders one component per Kind off the wire enum. A Kind added
// here and not to proto/tf2ap/launcher/v1/form.proto would reach it as the
// zero value and draw as nothing, which is why this compares both directions.
func TestProtoKindsMatchFormKinds(t *testing.T) {
	kinds := declaredKinds(t)
	values := launcherv1.Kind(0).Descriptor().Values()

	if got, want := values.Len()-1, len(kinds); got != want {
		t.Fatalf("proto declares %d kinds beside KIND_UNSPECIFIED, form declares %d", got, want)
	}
	if name := values.ByNumber(0).Name(); name != "KIND_UNSPECIFIED" {
		t.Errorf("proto kind 0 is %s, want KIND_UNSPECIFIED: form starts at one", name)
	}

	for _, kind := range kinds {
		value := values.ByNumber(launcherv1.Kind(kind).Number())
		if value == nil {
			t.Errorf("form kind %s is %d, which no proto kind holds", kind, kind)
			continue
		}
		if got := strings.ToLower(strings.TrimPrefix(string(value.Name()), "KIND_")); got != kind.String() {
			t.Errorf("proto kind %d is %s, form kind %d is %s", kind, value.Name(), kind, kind)
		}
	}
}
