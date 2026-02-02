package defaultregistry

import (
	"bytes"
	"strings"
	"time"

	"github.com/armosec/gojay"
)

func decodeV1Mafinset(raw []byte) (*shortV1Manifest, error) {
	manifest := shortV1Manifest{}
	if err := gojay.NewDecoder(bytes.NewReader(raw)).DecodeObject(&manifest); err != nil && !strings.Contains(err.Error(), "EOF") {
		return nil, err
	}
	return &manifest, nil
}

type shortV1Manifest struct {
	history history
}

func (m *shortV1Manifest) UnmarshalJSONObject(dec *gojay.Decoder, key string) (err error) {
	if key == "history" {
		m.history = history{}
		err = dec.DecodeArray(&m.history)
	}
	return err
}

func (r *shortV1Manifest) NKeys() int {
	return 1
}

type history []shortV1HistoryObj

func (h *history) UnmarshalJSONArray(dec *gojay.Decoder) error {
	v1HistoryObj := shortV1HistoryObj{}
	if err := dec.Object(&v1HistoryObj); err != nil {
		return err
	}
	*h = append(*h, v1HistoryObj)
	return nil
}

func (n *history) NKeys() int {
	return 0
}

type shortV1HistoryObj struct {
	v1Compatibility shortV1Compatibility
}

func (ho *shortV1HistoryObj) UnmarshalJSONObject(dec *gojay.Decoder, key string) (err error) {
	if key == "v1Compatibility" {
		var v1CompatibilityStr string
		err = dec.String(&v1CompatibilityStr)
		if err != nil {
			return err
		}
		v1Compatibility := shortV1Compatibility{}
		if err := gojay.NewDecoder(bytes.NewReader([]byte(v1CompatibilityStr))).DecodeObject(&v1Compatibility); err != nil && !strings.Contains(err.Error(), "EOF") {
			return err
		}
		ho.v1Compatibility = v1Compatibility
	}
	return err
}
func (r *shortV1HistoryObj) NKeys() int {
	return 1
}

type shortV1Compatibility struct {
	created time.Time
}

func (v1Comp *shortV1Compatibility) UnmarshalJSONObject(dec *gojay.Decoder, key string) (err error) {
	if key == "created" {
		var createdStr string
		err = dec.String(&createdStr)
		if err == nil {
			v1Comp.created, err = time.Parse(time.RFC3339, createdStr)
		}
	}
	return err
}

func (n *shortV1Compatibility) NKeys() int {
	return 1
}
