package chezmoi

// A PersistentState is a state peristed.
type PersistentState interface {
	Delete(bucket, key []byte) error
	Get(bucket, key []byte) ([]byte, error)
	Set(bucket, key, value []byte) error
}
