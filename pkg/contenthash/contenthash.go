// Package contenthash produces stable, byte-identical hashes for Go values
// destined for the session Merkle DAG.
//
// Canonical form:
//   - JSON, sorted object keys, no whitespace
//   - UTF-8 strings, NFC-normalized
//   - Null map values elided (treated as missing)
//   - Numbers preserved as-decoded; integer-shaped numbers are written without
//     fractional/exponent forms
//
// Hash inputs flow through Canonical -> SHA-256 -> lowercase hex.
package contenthash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// Canonical encodes v as canonical JSON.
func Canonical(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("contenthash: marshal: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var decoded any
	if err := dec.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("contenthash: decode: %w", err)
	}
	var buf bytes.Buffer
	if err := writeValue(&buf, decoded); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Hash returns the lowercase hex SHA-256 of Canonical(v).
func Hash(v any) (string, error) {
	b, err := Canonical(v)
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

// HashBytes returns the lowercase hex SHA-256 of b. Use when callers already
// have canonical bytes in hand.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func writeValue(buf *bytes.Buffer, v any) error {
	switch val := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case string:
		return writeString(buf, val)
	case json.Number:
		s := val.String()
		// Integer-shaped numbers stay integer-shaped.
		if !strings.ContainsAny(s, ".eE") {
			buf.WriteString(s)
			return nil
		}
		buf.WriteString(s)
	case []any:
		buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeValue(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(val))
		for k, kv := range val {
			if kv == nil {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		buf.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := writeString(buf, k); err != nil {
				return err
			}
			buf.WriteByte(':')
			if err := writeValue(buf, val[k]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	default:
		return fmt.Errorf("contenthash: unsupported type %T", v)
	}
	return nil
}

func writeString(buf *bytes.Buffer, s string) error {
	enc, err := json.Marshal(norm.NFC.String(s))
	if err != nil {
		return err
	}
	buf.Write(enc)
	return nil
}
