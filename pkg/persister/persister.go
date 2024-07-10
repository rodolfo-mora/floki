package persister

// Interface used to persist data to any
// long term storage engine.
type Persister interface {
	Set(key string, value interface{}) error
	Get(key string) (string, error)
	Delete(key string) error
}
