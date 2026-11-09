package yaml_test

import (
	"bytes"
	"cmp"
	_ "embed"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/iotest"

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

func TestMarshal_ambiguousStrings(t *testing.T) {
	in := map[string]any{
		"bool":      true,
		"boolStr":   "true",
		"null":      nil,
		"nullStr":   "null",
		"int":       42,
		"intStr":    "42",
		"float":     3.14,
		"floatStr":  "3.14",
		"empty":     "",
		"plain":     "hello",
		"intKeyStr": map[string]string{"1": "one"},
	}

	b, err := yaml.Marshal(in, json.Deterministic(true))
	if err != nil {
		t.Fatal(err)
	}

	const want = `bool: true
boolStr: "true"
empty: ""
float: 3.14
floatStr: "3.14"
int: 42
intKeyStr:
    "1": one
intStr: "42"
"null": null
nullStr: "null"
plain: hello
`
	if string(b) != want {
		t.Fatalf("got:\n%s\nwant:\n%s", b, want)
	}

	out := map[string]any{}
	if err := yaml.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	in["int"] = float64(42) // numbers decode into any as float64
	in["intKeyStr"] = map[string]any{"1": "one"}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("round trip: got %#v, want %#v", out, in)
	}
}

func TestMarshal_options(t *testing.T) {
	type tagged struct {
		Renamed   string `json:"renamed"`
		Omitted   string `json:"omitted,omitempty"`
		Skipped   string `json:"-"`
		Stringify int    `json:"stringify,string"`
	}

	type zeroable struct {
		A int    `json:"a"`
		B string `json:"b"`
		C []int  `json:"c"`
	}

	for _, tc := range []struct {
		name string
		in   any
		opts []json.Options
		want string
	}{
		{
			name: "struct tags",
			in:   tagged{Renamed: "x", Skipped: "y", Stringify: 7},
			want: "renamed: x\nstringify: \"7\"\n",
		},
		{
			name: "zero values by default",
			in:   zeroable{},
			want: "a: 0\nb: \"\"\nc: []\n",
		},
		{
			name: "OmitZeroStructFields",
			in:   zeroable{B: "set"},
			opts: []json.Options{json.OmitZeroStructFields(true)},
			want: "b: set\n",
		},
		{
			name: "FormatNilSliceAsNull",
			in:   zeroable{},
			opts: []json.Options{json.FormatNilSliceAsNull(true)},
			want: "a: 0\nb: \"\"\nc: null\n",
		},
		{
			name: "StringifyNumbers",
			in:   zeroable{A: 42},
			opts: []json.Options{json.StringifyNumbers(true)},
			want: "a: \"42\"\nb: \"\"\nc: []\n",
		},
		{
			name: "Deterministic",
			in:   map[string]int{"c": 3, "a": 1, "b": 2},
			opts: []json.Options{json.Deterministic(true)},
			want: "a: 1\nb: 2\nc: 3\n",
		},
		{
			name: "multiline string as literal block",
			in:   map[string]string{"text": "line one\nline two\n"},
			want: "text: |\n    line one\n    line two\n",
		},
		{
			name: "top-level list",
			in:   []any{1, "two", false},
			want: "- 1\n- two\n- false\n",
		},
		{
			name: "top-level scalar",
			in:   "true",
			want: "\"true\"\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := yaml.Marshal(tc.in, tc.opts...)
			if err != nil {
				t.Fatal(err)
			}

			if string(b) != tc.want {
				t.Fatalf("got:\n%q\nwant:\n%q", b, tc.want)
			}
		})
	}
}

type failingMarshaler struct{}

var errMarshal = errors.New("marshal failed")

func (failingMarshaler) MarshalJSON() ([]byte, error) { return nil, errMarshal }

func TestMarshal_error(t *testing.T) {
	if _, err := yaml.Marshal(make(chan int)); err == nil {
		t.Fatal("expected an error for an unsupported type")
	}

	if _, err := yaml.Marshal(failingMarshaler{}); !errors.Is(err, errMarshal) {
		t.Fatalf("got %v, want %v", err, errMarshal)
	}
}

func TestUnmarshal_files(t *testing.T) {
	for _, name := range []string{"example.yaml", "different order.yaml"} {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}

			got := testStruct{}
			if err := yaml.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, structExample) {
				t.Fatalf("got %#v, want %#v", got, structExample)
			}
		})
	}
}

func TestUnmarshal_json(t *testing.T) {
	b, err := os.ReadFile("example.json")
	if err != nil {
		t.Fatal(err)
	}

	got := testStruct{}
	if err := yaml.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}

	want := structExample
	want.Block += "\n"
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	// example.json has members testStruct lacks
	if err := yaml.Unmarshal(b, &testStruct{}, json.RejectUnknownMembers(true)); err == nil {
		t.Fatal("expected an error for unknown members")
	}
}

func TestUnmarshal_scalars(t *testing.T) {
	const in = `plain_int: 42
plain_float: 1.5
plain_bool: false
plain_null: null
single_quoted: '42'
double_quoted: "true"
literal: |
    keep
folded: >-
    one
    two
anchor: &a hello
alias: *a
`

	got := map[string]any{}
	if err := yaml.Unmarshal([]byte(in), &got); err != nil {
		t.Fatal(err)
	}

	want := map[string]any{
		"plain_int":     float64(42),
		"plain_float":   1.5,
		"plain_bool":    false,
		"plain_null":    nil,
		"single_quoted": "42",
		"double_quoted": "true",
		"literal":       "keep\n",
		"folded":        "one two",
		"anchor":        "hello",
		"alias":         "hello",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestUnmarshal_options(t *testing.T) {
	type person struct {
		Name string `json:"name"`
	}

	in := []byte("NAME: Ada\n")

	got := person{}
	if err := yaml.Unmarshal(in, &got); err != nil {
		t.Fatal(err)
	}

	if got.Name != "" {
		t.Fatalf("names match case-sensitively by default, got %q", got.Name)
	}

	if err := yaml.Unmarshal(in, &got, json.MatchCaseInsensitiveNames(true)); err != nil {
		t.Fatal(err)
	}

	if got.Name != "Ada" {
		t.Fatalf("got %q, want %q", got.Name, "Ada")
	}

	if err := yaml.Unmarshal(in, &person{}, json.RejectUnknownMembers(true)); err == nil {
		t.Fatal("expected an error for an unknown member")
	}
}

func TestUnmarshal_errors(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		out  any
	}{
		{"invalid yaml", "a: [1", &map[string]any{}},
		{"non-string key", "1: one", &map[string]any{}},
		{"duplicate key", "a: 1\na: 2", &map[string]any{}},
		{"type mismatch", "name: [1, 2]", &struct {
			Name string `json:"name"`
		}{}},
		{"non-pointer", "a: 1", map[string]any{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := yaml.Unmarshal([]byte(tc.in), tc.out); err == nil {
				t.Fatal("expected an error")
			}
		})
	}

	t.Run("empty input", func(t *testing.T) {
		if err := yaml.Unmarshal(nil, &map[string]any{}); !errors.Is(err, io.EOF) {
			t.Fatalf("got %v, want %v", err, io.EOF)
		}
	})
}

func TestUnmarshalRead(t *testing.T) {
	f, err := os.Open("example.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	got := testStruct{}
	if err := yaml.UnmarshalRead(f, &got); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, structExample) {
		t.Fatalf("got %#v, want %#v", got, structExample)
	}

	if err := yaml.UnmarshalRead(strings.NewReader("name: Ada"), &got, json.RejectUnknownMembers(true)); err == nil {
		t.Fatal("expected options to reach Unmarshal")
	}

	errRead := errors.New("read failed")
	if err := yaml.UnmarshalRead(iotest.ErrReader(errRead), &got); !errors.Is(err, errRead) {
		t.Fatalf("got %v, want %v", err, errRead)
	}
}
