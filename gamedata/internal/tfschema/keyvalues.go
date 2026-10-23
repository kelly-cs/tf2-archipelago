// Package tfschema reads the two game files the weapon tools work from: the
// item schema, which is Valve KeyValues, and the English localisation, which is
// UTF-16 KeyValues read as lines.
package tfschema

import (
	"fmt"
	"os"
	"strings"
)

// Object is one KeyValues block: keys in the order they were first written,
// and for a key written twice, the last value.
type Object struct {
	keys   []string
	values map[string]Value
}

// Value is a string or a nested block, never both.
type Value struct {
	Text  string
	Block *Object
}

// Keys is every key, in writing order.
func (o *Object) Keys() []string { return o.keys }

// Get is one value and whether the key exists.
func (o *Object) Get(key string) (Value, bool) {
	v, ok := o.values[key]
	return v, ok
}

// Text is a string value, and empty when the key is absent or a block.
func (o *Object) Text(key string) string {
	if v, ok := o.values[key]; ok {
		return v.Text
	}
	return ""
}

// Block is a nested block, and nil when the key is absent or a string.
func (o *Object) Block(key string) *Object {
	if v, ok := o.values[key]; ok {
		return v.Block
	}
	return nil
}

func (o *Object) set(key string, v Value) {
	if _, seen := o.values[key]; !seen {
		o.keys = append(o.keys, key)
	}
	o.values[key] = v
}

func newObject() *Object { return &Object{values: make(map[string]Value)} }

// ParseFile reads a KeyValues file with one root block, and returns that block.
func ParseFile(path string) (*Object, error) {
	body, err := os.ReadFile(path) //nolint:gosec // the path is a flag on a maintainer tool
	if err != nil {
		return nil, err
	}
	root, err := Parse(string(body))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return root, nil
}

// Parse reads KeyValues text: quoted keys and values, braces for blocks, and
// nothing else. Whatever is not one of those, comments included, is skipped.
func Parse(text string) (*Object, error) {
	tokens := tokenize(strings.TrimPrefix(text, "\ufeff"))
	if len(tokens) < 2 || tokens[1].brace != '{' {
		return nil, fmt.Errorf("expected a root KeyValues object")
	}
	root, end, err := object(tokens, 2)
	if err != nil {
		return nil, err
	}
	if end != len(tokens) {
		return nil, fmt.Errorf("trailing KeyValues tokens")
	}
	return root, nil
}

type token struct {
	text  string
	brace byte
}

func tokenize(text string) []token {
	var tokens []token
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '{', '}':
			tokens = append(tokens, token{brace: text[i]})
		case '"':
			var b strings.Builder
			for i++; i < len(text) && text[i] != '"'; i++ {
				if text[i] == '\\' && i+1 < len(text) {
					i++
					switch text[i] {
					case 'n':
						b.WriteByte('\n')
					case 't':
						b.WriteByte('\t')
					case '"', '\\':
						b.WriteByte(text[i])
					default:
						b.WriteByte('\\')
						b.WriteByte(text[i])
					}
					continue
				}
				b.WriteByte(text[i])
			}
			tokens = append(tokens, token{text: b.String()})
		}
	}
	return tokens
}

func object(tokens []token, i int) (*Object, int, error) {
	out := newObject()
	for i < len(tokens) && tokens[i].brace != '}' {
		if i+1 >= len(tokens) {
			return nil, 0, fmt.Errorf("a key with no value at the end of the file")
		}
		key, value := tokens[i], tokens[i+1]
		i += 2
		switch value.brace {
		case '{':
			block, end, err := object(tokens, i)
			if err != nil {
				return nil, 0, err
			}
			out.set(key.text, Value{Block: block})
			i = end
		case '}':
			return nil, 0, fmt.Errorf("the key %q has no value", key.text)
		default:
			out.set(key.text, Value{Text: value.text})
		}
	}
	return out, i + 1, nil
}
