package file_readers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// DecodeJSON decodes JSON data from the reader into the provided structure
func DecodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, JSON, err)
	}
	return nil
}

// DecodeYAML decodes YAML data from the reader into the provided structure
func DecodeYAML(r io.Reader, v any) error {
	if err := yaml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, YAML, err)
	}
	return nil
}

// DecodeTOML decodes TOML data from the reader into the provided structure
func DecodeTOML(r io.Reader, v any) error {
	if _, err := toml.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf(decodeFailedTemplate, TOML, err)
	}
	return nil
}

// DecodeCSV decodes CSV data from the reader into the provided structure
func DecodeCSV(r io.Reader, v any) error {
	err := NewCSVDecoder(r).Decode(v)
	if err == io.EOF{
		return io.EOF
	} 
	if err != nil {
		return fmt.Errorf(decodeFailedTemplate, CSV, err)
	}
	return nil
}

type CSVDecoder struct {
	r *csv.Reader
	header []string
	headerLoaded bool
}

func NewCSVDecoder(r io.Reader) *CSVDecoder {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	return &CSVDecoder{
		r: cr,
	}
}

func (d *CSVDecoder) Decode(v any) error {
	dst, ok := v.(*map[string]string)
	if !ok {
		return fmt.Errorf("CSVMapDecoder: expected *map[string]string, got %T", v)
	}
	if !d.headerLoaded {
		header, err := d.r.Read()
		if err != nil {
			return fmt.Errorf("CSVMapDecoder: cannot read csv header: %w", err)
		}
		d.header = header
		d.headerLoaded = true
	}
	row, err := d.r.Read()
	if err != nil {
		if err == csv.ErrFieldCount {
			return fmt.Errorf("CSVMapDecoder: invalid field count: %w", err)
		}
		return err
	}
	record := make(map[string]string, len(d.header))
	n := min(len(d.header), len(row))
	
	for ind := range n {
		record[d.header[ind]] = row[ind]
	}
	for ind := n; ind < len(d.header); ind++ {
		record[d.header[ind]] = ""
	} 
	*dst = record
	return nil
}