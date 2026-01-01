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
	dst, ok := v.(*[]map[string]string)
	if !ok {
		return fmt.Errorf("DecodeCSV: expected *[]map[string]string, got %T", v)
	}
	decoder := NewCSVDecoder(r)
	for {
		var record map[string]string
		err := decoder.Decode(&record)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf(decodeFailedTemplate, CSV, err)
		}
		*dst = append(*dst, record)
	}
	return nil
}

// GetDecoderByExtension returns the appropriate decoder function for the given file extension
func GetDecoderByExtension(extension string) (DecodeFunc, error) {
	switch extension {
	case "json":
		return DecodeJSON, nil
	case "yaml", "yml":
		return DecodeYAML, nil
	case "toml":
		return DecodeTOML, nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", extension)
	}
}

// CSVDecoder is a simple CSV decoder that reads records into map[string]string
type CSVDecoder struct {
	r            *csv.Reader
	header       []string
	headerLoaded bool
}

// NewCSVDecoder creates a new CSVDecoder instance
func NewCSVDecoder(r io.Reader) *CSVDecoder {
	cr := csv.NewReader(r)
	cr.TrimLeadingSpace = true
	return &CSVDecoder{
		r: cr,
	}
}

// Decode reads the next CSV record and decodes it into the provided map
func (d *CSVDecoder) Decode(v any) error {
	dst, ok := v.(*map[string]string)
	if !ok {
		return fmt.Errorf("CSVDecoder: expected *map[string]string, got %T", v)
	}
	if !d.headerLoaded {
		header, err := d.r.Read()
		if err == io.EOF {
			return io.EOF
		}
		if err != nil {
			return fmt.Errorf("CSVDecoder: cannot read csv header: %w", err)
		}
		d.header = header
		d.headerLoaded = true
	}
	row, err := d.r.Read()
	if err != nil {
		if err == csv.ErrFieldCount {
			return fmt.Errorf("CSVDecoder: invalid field count: %w", err)
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
