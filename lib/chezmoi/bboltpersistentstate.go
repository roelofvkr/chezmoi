package chezmoi

import (
	vfs "github.com/twpayne/go-vfs"
	"go.etcd.io/bbolt"
)

// A BBoltPersistentState is a state persisted with bbolt.
type BBoltPersistentState struct {
	*bbolt.DB
}

// NewBBoltPersistentState returns a new BBoltPersistentState.
func NewBBoltPersistentState(fs vfs.FS, path string) (*BBoltPersistentState, error) {
	return BBoltPersistentState{
		DB: bbolt.Open(path, 0600, &bbolt.Options{
			OpenFile: fs.OpenFile,
		}),
	}, nil
}
