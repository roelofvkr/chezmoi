package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestScriptState(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.config/chezmoi": &vfst.Dir{Perm: 0755},
	})
	require.NoError(t, err)
	defer cleanup()

	s, err := NewScriptState(fs, "/home/user/.config/chezmoi/script-state.yaml", 0600)
	require.NoError(t, err)

	script := []byte("#/bin/sh\ntrue\n")
	ran, err := s.GetScriptRanState(script)
	assert.NoError(t, err)
	assert.False(t, ran)

	assert.NoError(t, s.SetScriptRanState(script))

	ran, err = s.GetScriptRanState(script)
	assert.NoError(t, err)
	assert.True(t, ran)

	s2, err := NewScriptState(fs, "/home/user/.config/chezmoi/script-state.yaml", 0600)
	require.NoError(t, err)
	ran, err = s2.GetScriptRanState(script)
	assert.NoError(t, err)
	assert.True(t, ran)
}
