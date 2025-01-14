package yaml_test

import (
	"bytes"
	_ "embed"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/MarkRosemaker/yaml"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
)

var (
	//go:embed example.json
	exampleJSON jsontext.Value
	//go:embed example.yaml
	exampleYAML []byte
)

type testStruct struct {
	String     string            `json:"string"`
	Int        int               `json:"int"`
	Float      float64           `json:"float"`
	Bool       bool              `json:"bool"`
	Bool2      bool              `json:"bool2"`
	Null       *string           `json:"null_value"`
	List       []string          `json:"list"`
	Dictionary map[string]string `json:"dictionary"`
	Nested     nested            `json:"nested"`
	Block      string            `json:"block"`
}

type nested struct {
	ListOfDicts []dictionary        `json:"list_of_dicts"`
	DictOfLists map[string][]string `json:"dict_of_lists"`
}

type dictionary struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestMarshal(t *testing.T) {
	res, err := yaml.Marshal(testStruct{
		String: "Hello, World!",
		Int:    42,
		Float:  3.14,
		Bool:   true,
		Bool2:  false,
		Null:   nil,
		List: []string{
			"item1",
			"item2",
			"item3",
		},
		Dictionary: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
		Nested: nested{
			ListOfDicts: []dictionary{
				{
					Name:  "item1",
					Value: 1,
				},
				{
					Name:  "item2",
					Value: 2,
				},
			},
			DictOfLists: map[string][]string{
				"key1": {
					"item1",
					"item2",
				},
				"key2": {
					"item3",
					"item4",
				},
			},
		},
		Block: `This is a block
style multiline string.`,
	}, json.WithMarshalers(json.NewMarshalers(json.MarshalFuncV2(
		func(enc *jsontext.Encoder, m map[string]string, opts json.Options) error {

			if err := enc.WriteToken(jsontext.ObjectStart); err != nil {
				return err
			}

			keys := slices.Collect(maps.Keys(m))
			slices.Sort(keys)

			for _, k := range keys {
				if err := enc.WriteToken(jsontext.String(k)); err != nil {
					return err
				}

				if err := enc.WriteToken(jsontext.String(m[k])); err != nil {
					return err
				}
			}

			return enc.WriteToken(jsontext.ObjectEnd)
		}))))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res, exampleYAML) {
		os.WriteFile("real.yaml", res, 0644)
		t.Fatalf("got: %q, want: %q", res, exampleYAML)
	}
}
