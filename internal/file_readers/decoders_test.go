package file_readers

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeJSON_Success(t *testing.T) {
	type payload struct {
		A int    `json:"a"`
		B string `json:"b"`
	}

	var got payload
	err := DecodeJSON(strings.NewReader(`{"a": 1, "b": "x"}`), &got)
	require.NoError(t, err)
	require.Equal(t, 1, got.A)
	require.Equal(t, "x", got.B)
}

func TestDecodeJSON_Error(t *testing.T) {
	var got map[string]any
	err := DecodeJSON(strings.NewReader(`{"a":`), &got)
	require.Error(t, err)
}

func TestDecodeYAML_Success(t *testing.T) {
	type payload struct {
		A int    `yaml:"a"`
		B string `yaml:"b"`
	}

	var got payload
	err := DecodeYAML(strings.NewReader("a: 2\nb: y\n"), &got)
	require.NoError(t, err)
	require.Equal(t, 2, got.A)
	require.Equal(t, "y", got.B)
}

func TestDecodeYAML_Error(t *testing.T) {
	var got map[string]any
	err := DecodeYAML(strings.NewReader("a: [1,2\n"), &got) 
	require.Error(t, err)
}

func TestDecodeTOML_Success(t *testing.T) {
	type payload struct {
		A int    `toml:"a"`
		B string `toml:"b"`
	}

	var got payload
	err := DecodeTOML(strings.NewReader("a = 3\nb = \"z\"\n"), &got)
	require.NoError(t, err)
	require.Equal(t, 3, got.A)
	require.Equal(t, "z", got.B)
}

func TestDecodeTOML_Error(t *testing.T) {
	var got map[string]any
	err := DecodeTOML(strings.NewReader("a = \n"), &got)
	require.Error(t, err)
}

func TestDecodeCSV_Success(t *testing.T) {
	var dst []map[string]string
	csv := "a,b,c\n1,2,3\n4,5,6\n"

	err := DecodeCSV(strings.NewReader(csv), &dst)
	require.NoError(t, err)

	require.Len(t, dst, 2)
	require.Equal(t, map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}, dst[0])
	require.Equal(t, map[string]string{
		"a": "4",
		"b": "5",
		"c": "6",
	}, dst[1])
}

func TestDecodeCSV_WrongType_ReturnsError(t *testing.T) {
	var notSlice map[string]string
	err := DecodeCSV(strings.NewReader("a,b\n1,2\n"), &notSlice)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected *[]map[string]string")
}

func TestCSVDecoder_Decode_EmptyReader_ReturnsEOF(t *testing.T) {
	dec := NewCSVDecoder(strings.NewReader(""))
	var rec map[string]string
	err := dec.Decode(&rec)
	require.ErrorIs(t, err, io.EOF)
}

func TestCSVDecoder_Decode_HeaderReadError(t *testing.T) {
	dec := NewCSVDecoder(strings.NewReader("a,\"b\n1,2\n"))
	var rec map[string]string
	err := dec.Decode(&rec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot read csv header")
}

func TestCSVDecoder_Decode_InvalidFieldCount_Wrapped(t *testing.T) {
	dec := NewCSVDecoder(strings.NewReader("a,b,c\n1,2\n")) 
	var rec map[string]string
	err := dec.Decode(&rec)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong number of fields")
}

func TestCSVDecoder_Decode_FewerColumns_FillsMissingWithEmpty_WhenVariableFieldsAllowed(t *testing.T) {
	dec := NewCSVDecoder(strings.NewReader("a,b,c\n1,2\n"))
	dec.r.FieldsPerRecord = -1 

	var rec map[string]string
	err := dec.Decode(&rec)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"a": "1",
		"b": "2",
		"c": "",
	}, rec)
}

func TestCSVDecoder_Decode_MoreColumns_IgnoresExtra_WhenVariableFieldsAllowed(t *testing.T) {
	dec := NewCSVDecoder(strings.NewReader("a,b\n1,2,3,4\n"))
	dec.r.FieldsPerRecord = -1

	var rec map[string]string
	err := dec.Decode(&rec)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"a": "1",
		"b": "2",
	}, rec)
}

func TestCSVDecoder_TrimLeadingSpace_TrimsValues(t *testing.T) {
	dec := NewCSVDecoder(strings.NewReader("a,b\n  x,   y\n"))
	dec.r.FieldsPerRecord = -1

	var rec map[string]string
	err := dec.Decode(&rec)
	require.NoError(t, err)

	require.Equal(t, "x", rec["a"])
	require.Equal(t, "y", rec["b"])
}

func TestDecodeCSV_StopsOnEOF_NoError(t *testing.T) {
	var dst []map[string]string
	err := DecodeCSV(strings.NewReader("a,b\n1,2\n"), &dst)
	require.NoError(t, err)
	require.Len(t, dst, 1)
}

func TestDecodeCSV_ReturnsDecodeFailedTemplateWrappedOnDecoderError(t *testing.T) {
	var dst []map[string]string
	err := DecodeCSV(strings.NewReader("a,b,c\n1,2\n"), &dst)
	require.Error(t, err)

	require.True(t, errors.Is(err, io.EOF) == false)
	require.Contains(t, err.Error(), "wrong number of fields")
}

func TestGetDecoderByExtension_JSON(t *testing.T) {
	decoder, err := GetDecoderByExtension("json")
	require.NoError(t, err)
	require.NotNil(t, decoder)
}

func TestGetDecoderByExtension_YAML(t *testing.T) {
	decoder, err := GetDecoderByExtension("yaml")
	require.NoError(t, err)
	require.NotNil(t, decoder)

	decoder, err = GetDecoderByExtension("yml")
	require.NoError(t, err)
	require.NotNil(t, decoder)
}

func TestGetDecoderByExtension_TOML(t *testing.T) {
	decoder, err := GetDecoderByExtension("toml")
	require.NoError(t, err)
	require.NotNil(t, decoder)
}

func TestGetDecoderByExtension_Unsupported(t *testing.T) {
	decoder, err := GetDecoderByExtension("csv")
	require.Error(t, err)
	require.Nil(t, decoder)
	require.Contains(t, err.Error(), "unsupported file extension")

	decoder, err = GetDecoderByExtension("xml")
	require.Error(t, err)
	require.Nil(t, decoder)
}
