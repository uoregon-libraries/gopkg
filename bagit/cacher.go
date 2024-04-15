package bagit

// The Cacher interface defines a simple way for a bag's manifest data to
// optionally be cached or pre-computed by the caller when building manifests.
// Validation functions do not use this.
type Cacher interface {
	GetSum(path string) (value string, exists bool)
	SetSum(path, value string)
}

type noopCache struct{}

func (noopCache) GetSum(_ string) (string, bool) {
	return "", false
}

func (noopCache) SetSum(_, _ string) {
}
