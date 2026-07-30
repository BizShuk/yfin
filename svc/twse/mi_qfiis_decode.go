// mi_qfiis_decode.go — provider-wire decoding for MI_QFIIS rows, whose
// percentage cells are JSON numbers while the public TWSE response contract
// represents every cell as a string.
package twse

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bizshuk/yfin/model"
)

type miQFIISWireResponse struct {
	Stat   string              `json:"stat"`
	Title  string              `json:"title,omitempty"`
	Fields []string            `json:"fields"`
	Data   [][]json.RawMessage `json:"data"`
	Notes  string              `json:"notes,omitempty"`
	Total  int                 `json:"total,omitempty"`
	Date   string              `json:"date"`
}

func (r *miQFIISWireResponse) GetStat() string {
	return r.Stat
}

func (r *miQFIISWireResponse) toModel() (model.MI_QFIISResponse, error) {
	var notes []string
	if r.Notes != "" {
		notes = []string{r.Notes}
	}

	rows := make([][]string, len(r.Data))
	for rowIndex, rawRow := range r.Data {
		row := make([]string, len(rawRow))
		for columnIndex, rawCell := range rawRow {
			cell, err := decodeMIQFIISCell(rawCell)
			if err != nil {
				return model.MI_QFIISResponse{}, fmt.Errorf(
					"MI_QFIIS: decode row %d column %d: %w",
					rowIndex,
					columnIndex,
					err,
				)
			}
			row[columnIndex] = cell
		}
		rows[rowIndex] = row
	}

	return model.MI_QFIISResponse{
		Response: model.Response{
			Stat:   r.Stat,
			Title:  r.Title,
			Fields: r.Fields,
			Data:   rows,
			Notes:  notes,
			Total:  r.Total,
		},
		Date: r.Date,
	}, nil
}

func decodeMIQFIISCell(raw json.RawMessage) (string, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return "", fmt.Errorf("decode JSON value: %w", err)
	}

	switch value := value.(type) {
	case string:
		return value, nil
	case json.Number:
		return value.String(), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported JSON value type %T", value)
	}
}
