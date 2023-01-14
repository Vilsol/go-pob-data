package loader

type AssetCache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Exists(key string) bool
}
