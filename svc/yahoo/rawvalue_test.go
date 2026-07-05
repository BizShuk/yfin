package yahoo

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawValue_DecodesRawAndFmt(t *testing.T) {
	var v RawValue
	require.NoError(t, json.Unmarshal([]byte(`{"raw":123.45,"fmt":"123.45","longFmt":"123"}`), &v))
	require.NotNil(t, v.Raw)
	require.Equal(t, 123.45, *v.Raw)
	require.Equal(t, "123.45", v.Fmt)
}

func TestRawValue_HandlesEmptyObject(t *testing.T) {
	var v RawValue
	require.NoError(t, json.Unmarshal([]byte(`{}`), &v))
	require.Nil(t, v.Raw)
}

func TestRawInt_DecodesRaw(t *testing.T) {
	var v RawInt
	require.NoError(t, json.Unmarshal([]byte(`{"raw":1700000000,"fmt":"..."}`), &v))
	require.NotNil(t, v.Raw)
	require.Equal(t, int64(1700000000), *v.Raw)
}
