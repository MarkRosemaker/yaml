package yaml

import (
	"github.com/MarkRosemaker/json2yaml"
	"github.com/MarkRosemaker/yaml2json"
	"github.com/go-json-experiment/json"
	"gopkg.in/yaml.v3"
)

// Marshal serializes a Go value as a []byte according to the provided
// marshal and encode options (while ignoring unmarshal or decode options).
func Marshal(in any, opts ...json.Options) ([]byte, error) {
	// marshal as json with the given options
	out, err := json.Marshal(in, opts...)
	if err != nil {
		return nil, err
	}

	// convert to yaml
	n, err := json2yaml.Convert(out)
	if err != nil {
		return nil, err
	}

	// marshal the yaml to bytes
	return yaml.Marshal(n)
}

// Unmarshal decodes a []byte input into a Go value according to the provided
// unmarshal and decode options (while ignoring marshal or encode options).
// The output must be a non-nil pointer.
func Unmarshal(in []byte, out any, opts ...json.Options) error {
	// parse input into a yaml document node
	n := &yaml.Node{}
	if err := yaml.Unmarshal(in, n); err != nil {
		return err
	}

	// convert the yaml to json
	val, err := yaml2json.Convert(n)
	if err != nil {
		return err
	}

	return json.Unmarshal(val, out, opts...)
}
