package chezmoi

import (
	"crypto/sha256"
	"encoding/hex"
	"os"

	vfs "github.com/twpayne/go-vfs"
	"gopkg.in/yaml.v2"
)

// FIXME use a proper database

// A ScriptState persists state about scripts have been run.
type ScriptState struct {
	fs       vfs.FS
	filename string
	perm     os.FileMode
	cache    map[string]bool
}

// NewScriptState returns a new ScriptState.
func NewScriptState(fs vfs.FS, filename string, perm os.FileMode) (*ScriptState, error) {
	cache := make(map[string]bool)
	data, err := fs.ReadFile(filename)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, &cache); err != nil {
			return nil, err
		}
	case !os.IsNotExist(err):
		return nil, err
	}
	return &ScriptState{
		fs:       fs,
		filename: filename,
		perm:     perm,
		cache:    cache,
	}, nil
}

// GetScriptRanState returns if script has run.
func (s *ScriptState) GetScriptRanState(script []byte) (bool, error) {
	return s.cache[s.getKey(script)], nil
}

// SetScriptRanState sets the run state of script.
func (s *ScriptState) SetScriptRanState(script []byte) error {
	s.cache[s.getKey(script)] = true
	data, err := yaml.Marshal(s.cache)
	if err != nil {
		return err
	}
	return s.fs.WriteFile(s.filename, data, s.perm)
}

func (s *ScriptState) getKey(script []byte) string {
	hashArr := sha256.Sum256(script)
	return hex.EncodeToString(hashArr[:])
}
