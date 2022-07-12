package harbor

import (
	"bytes"
	"strings"

	"github.com/armosec/gojay"
)

type keyDecoder struct {
	key   string
	value string
}

func (k *keyDecoder) UnmarshalJSONObject(dec *gojay.Decoder, key string) (err error) {

	if key == k.key {
		err = dec.String(&(k.value))
	}
	return err
}

func (r *keyDecoder) NKeys() int {
	return 1
}

type names []string

func (n *names) UnmarshalJSONArray(dec *gojay.Decoder) error {
	d := keyDecoder{key: "name"}
	if err := dec.Object(&d); err != nil {
		return err
	}
	*n = append(*n, d.value)
	return nil
}
func (n *names) NKeys() int {
	return 0
}

func decodeObjectsNames(raw []byte) ([]string, error) {
	n := names{}
	if err := gojay.NewDecoder(bytes.NewReader(raw)).DecodeArray(&n); err != nil && !strings.Contains(err.Error(), "EOF") {
		return nil, err
	}
	return n, nil
}
