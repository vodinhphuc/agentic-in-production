package adapters

import (
	"errors"
	"sync"
)

// Factory builds an adapter from per-instance configuration JSON.
type Factory func(configJSON []byte) (AgentPlatformAdapter, error)

var (
	regMu     sync.RWMutex
	factories = map[string]Factory{} // keyed by adapter_kind, e.g. "mock"
)

func Register(kind string, f Factory) {
	regMu.Lock()
	defer regMu.Unlock()
	factories[kind] = f
}

func Build(kind string, configJSON []byte) (AgentPlatformAdapter, error) {
	regMu.RLock()
	defer regMu.RUnlock()
	f, ok := factories[kind]
	if !ok {
		return nil, errors.New("unknown adapter kind: " + kind)
	}
	return f(configJSON)
}
