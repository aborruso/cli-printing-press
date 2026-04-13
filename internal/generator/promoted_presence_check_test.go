package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mvanhorn/cli-printing-press/internal/spec"
	"github.com/stretchr/testify/require"
)

// TestPromotedPresenceCheckUsesPromotedType guards against a class of
// generator bugs where a flag declared with one Go type was compared
// against the zero value of a different type.
//
// Background: ID-like int parameters (steamid, appid, etc.) are promoted
// to string at declaration by goTypeForParam / cobraFlagFuncForParam to
// avoid overflow and empty-vs-unset confusion. Before this fix, the
// promoted command template used zeroVal(p.Type) for the presence check —
// which returned "0" for the original int type, producing
//
//	if flagSteamid != 0 { ... }   // flagSteamid is string!
//
// go vet rejected the generated code with "mismatched types string and
// untyped int". This test asserts both that the emitted check uses the
// string zero and that go vet passes end-to-end on a spec shaped like
// steam-web's failing case.
func TestPromotedPresenceCheckUsesPromotedType(t *testing.T) {
	t.Parallel()

	apiSpec := minimalSpec("steam-like")
	// Replace the default items/list endpoint with one that carries an
	// ID-like int query parameter. The promoted-command codepath emits
	// the buggy presence check when such a flag is present.
	apiSpec.Resources["items"] = spec.Resource{
		Description: "Items",
		Endpoints: map[string]spec.Endpoint{
			"get": {
				Method:      "GET",
				Path:        "/items/get",
				Description: "Get items for a Steam account",
				Params: []spec.Param{
					{Name: "steamid", Type: "int", Required: false, Description: "Steam ID (64-bit)"},
				},
			},
		},
	}

	outputDir := filepath.Join(t.TempDir(), "steam-like-pp-cli")
	require.NoError(t, New(apiSpec, outputDir).Generate())

	// 1. Inspect the rendered promoted-command source directly — pinpoints
	//    the template emission even if go vet is unavailable.
	promotedFiles, err := filepath.Glob(filepath.Join(outputDir, "internal", "cli", "promoted_*.go"))
	require.NoError(t, err)
	require.NotEmpty(t, promotedFiles, "expected at least one promoted command file")

	var combined strings.Builder
	for _, f := range promotedFiles {
		data, err := os.ReadFile(f)
		require.NoError(t, err)
		combined.Write(data)
	}
	src := combined.String()

	require.Contains(t, src, `flagSteamid != ""`,
		"flag declared as string (ID promotion) must be compared against string zero")
	require.NotContains(t, src, "flagSteamid != 0",
		"flag declared as string must not be compared against int zero — this is the #189 bug")

	// 2. go vet catches the type mismatch at the Go toolchain level. This
	//    is the acceptance bar — if it passes, the generated code compiles.
	require.NoError(t, runGoVet(t, outputDir),
		"generated promoted command with ID-like int param must pass go vet")
}
