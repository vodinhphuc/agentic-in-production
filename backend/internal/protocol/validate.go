package protocol

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed events.schema.embed.json
var eventsSchemaBytes []byte

var (
	once     sync.Once
	compiled *jsonschema.Schema
	loadErr  error
)

func compileOnce() {
	once.Do(func() {
		c := jsonschema.NewCompiler()
		if err := c.AddResource("events.schema.json", bytes.NewReader(eventsSchemaBytes)); err != nil {
			loadErr = fmt.Errorf("embedded schema: %w", err)
			return
		}
		s, err := c.Compile("events.schema.json")
		if err != nil {
			loadErr = err
			return
		}
		compiled = s
	})
}

// ValidateEvent returns nil if b is a valid canonical agent event.
func ValidateEvent(b []byte) error {
	compileOnce()
	if loadErr != nil {
		return loadErr
	}
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("event not valid JSON: %w", err)
	}
	return compiled.Validate(v)
}
