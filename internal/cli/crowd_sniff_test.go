package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrowdSniffCmd_MissingAPIFlag(t *testing.T) {
	t.Parallel()
	cmd := newCrowdSniffCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required flag")
}

func TestCrowdSniffCmd_HelpOutput(t *testing.T) {
	t.Parallel()
	cmd := newCrowdSniffCmd()
	cmd.SetArgs([]string{"--help"})
	// --help causes Execute to return nil (prints help)
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestValidateCrowdSniffAPIName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		api     string
		wantErr string
	}{
		{name: "valid name", api: "notion", wantErr: ""},
		{name: "valid domain", api: "api.notion.com", wantErr: ""},
		{name: "valid URL", api: "https://api.notion.com/v1", wantErr: ""},
		{name: "empty", api: "", wantErr: "required"},
		{name: "whitespace only", api: "   ", wantErr: "required"},
		{name: "newline injection", api: "notion\nHost: evil.com", wantErr: "invalid characters"},
		{name: "null byte", api: "notion\x00evil", wantErr: "invalid characters"},
		{name: "path traversal", api: "../../.ssh/evil", wantErr: "path traversal"},
		{name: "backslash traversal", api: `..\..\evil`, wantErr: "path traversal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateCrowdSniffAPIName(tt.api)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestDefaultCrowdSniffCachePath(t *testing.T) {
	t.Parallel()

	path := defaultCrowdSniffCachePath("notion")
	assert.Contains(t, path, "crowd-sniff")
	assert.Contains(t, path, "notion-spec.yaml")
}

func TestIsHTTPS(t *testing.T) {
	t.Parallel()

	assert.True(t, isHTTPS("https://api.example.com"))
	assert.True(t, isHTTPS("HTTPS://API.EXAMPLE.COM"))
	assert.False(t, isHTTPS("http://api.example.com"))
	assert.False(t, isHTTPS("ftp://api.example.com"))
	assert.False(t, isHTTPS(""))
	assert.False(t, isHTTPS("not-a-url"))
}
