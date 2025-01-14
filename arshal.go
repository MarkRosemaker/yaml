package yaml

import (
	"github.com/MarkRosemaker/json2yaml"
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
