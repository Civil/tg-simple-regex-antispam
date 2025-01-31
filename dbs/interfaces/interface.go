package interfaces

type SharedDB interface {
	StoreValue(key []byte, value []byte) error
	LoadValue(key []byte) ([]byte, error)
}

type SharedDBInitFunc func(map[string]any) (SharedDB, error)
