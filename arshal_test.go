package yaml_test

import (
	"bytes"
	"cmp"
	_ "embed"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"maps"
	"slices"
	"testing"

	"github.com/MarkRosemaker/yaml"
)

//go:embed example.yaml
var exampleYAML []byte

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

var structExample = testStruct{
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
	res, err := yaml.Marshal(structExample,
		json.WithMarshalers(json.JoinMarshalers(
			json.MarshalToFunc(orderMap[string, string]),
			json.MarshalToFunc(orderMap[string, []string]),
		)))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(res, exampleYAML) {
		t.Fatalf("got: %q, want: %q", res, exampleYAML)
	}
}

func TestUnmarshal(t *testing.T) {
	ex := &testStruct{}
	if err := yaml.Unmarshal(exampleYAML, ex); err != nil {
		t.Fatal(err)
	}

	if ex.Float != structExample.Float {
		t.Fatalf("got: %f, want: %f", ex.Float, structExample.Float)
	}

	if ex.Int != structExample.Int {
		t.Fatalf("got: %d, want: %d", ex.Int, structExample.Int)
	}
}

func orderMap[K cmp.Ordered, V any](enc *jsontext.Encoder, m map[K]V) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}

	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)

	for _, k := range keys {
		if err := json.MarshalEncode(enc, k); err != nil {
			return err
		}

		if err := json.MarshalEncode(enc, m[k]); err != nil {
			return err
		}
	}

	return enc.WriteToken(jsontext.EndObject)
}
