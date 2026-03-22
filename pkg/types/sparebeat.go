package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

type SparebeatMap struct {
	ID        string         `json:"id,omitempty"`
	Title     string         `json:"title"`
	Artist    string         `json:"artist"`
	URL       string         `json:"url"`
	BgColor   []string       `json:"bgColor,omitempty"`
	Beats     uint           `json:"beats,omitempty"`
	BPM       NumberOrString `json:"bpm"`
	StartTime int            `json:"startTime"`
	Level     Level          `json:"level"`
	Map       MapData        `json:"map"`
}

type Level struct {
	Easy   NumberOrString `json:"easy"`
	Normal NumberOrString `json:"normal"`
	Hard   NumberOrString `json:"hard"`
}

type MapEntry struct {
	Section string
	Options json.RawMessage
}

func (e *MapEntry) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		e.Section = str
		return nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("expected a section string or options object, got: %s", data)
	}
	e.Options = json.RawMessage(data)
	return nil
}

type MapData struct {
	Easy   []MapEntry `json:"easy"`
	Normal []MapEntry `json:"normal"`
	Hard   []MapEntry `json:"hard"`
}

type MapOptions struct {
	BarLine *bool    `json:"barLine,omitempty"`
	BPM     *float64 `json:"bpm,omitempty"`
	Speed   *float64 `json:"speed,omitempty"`
}

type NumberOrString struct {
	Value interface{}
}

func (v *NumberOrString) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch raw.(type) {
	case float64, string, nil:
		v.Value = raw
	default:
		return fmt.Errorf("expected a number or string, got: %s", data)
	}
	return nil
}

func ParseSparebeatMap(data []byte) (SparebeatMap, error) {
	var m SparebeatMap

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return m, err
	}
	if dec.More() {
		return m, errors.New(`unexpected data after JSON object`)
	}
	return m, m.validate()
}

func ParseMapOptions(raw json.RawMessage) (MapOptions, error) {
	var opts MapOptions
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&opts); err != nil {
		return opts, fmt.Errorf("invalid map options %s: %w", raw, err)
	}
	return opts, nil
}

func (m SparebeatMap) validate() error {
	if m.Title == "" {
		return errors.New(`missing or empty required field "title"`)
	}
	if m.Artist == "" {
		return errors.New(`missing or empty required field "artist"`)
	}
	if m.BPM.Value == nil {
		return errors.New(`missing required field "bpm"`)
	}
	return nil
}
