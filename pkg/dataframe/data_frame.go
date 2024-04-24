package dataframe

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boilingdata/go-boilingdata/models"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func initiateNewDataFrame(refID string, firstDataEntry map[string]interface{}) *data.Frame {
	frame := data.NewFrame("response")
	// add fields
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{}),
	)
	frame.RefID = refID
	var field *data.Field
	for key, value := range firstDataEntry {
		switch value.(type) {
		case int8:
			field = data.NewField(key, nil, []int8{})
		case int16:
			field = data.NewField(key, nil, []string{})
		case int32:
			field = data.NewField(key, nil, []int32{})
		case int64:
			field = data.NewField(key, nil, []int64{})
		case uint8:
			field = data.NewField(key, nil, []uint8{})
		case uint16:
			field = data.NewField(key, nil, []uint16{})
		case uint32:
			field = data.NewField(key, nil, []uint32{})
		case uint64:
			field = data.NewField(key, nil, []uint64{})
		case float32:
			field = data.NewField(key, nil, []float32{})
		case float64:
			field = data.NewField(key, nil, []float64{})
		case string:
			field = data.NewField(key, nil, []string{})
		case bool:
			field = data.NewField(key, nil, []bool{})
		case time.Time:
			field = data.NewField(key, nil, []time.Time{})
		case json.RawMessage:
			field = data.NewField(key, nil, []json.RawMessage{})
		case data.EnumItemIndex:
			field = data.NewField(key, nil, []data.EnumItemIndex{})
		default:
			panic(fmt.Errorf("value '%s' specified with unsupported type", value))
		}
		frame.Fields = append(frame.Fields, field)
	}
	return frame
}

func GetFrames(refID string, response *models.Response) *data.Frame {
	var frame *data.Frame
	if response != nil && len(response.Data) > 0 {
		frame = initiateNewDataFrame(refID, response.Data[0])
		for _, dataItem := range response.Data {
			vals := make([]interface{}, len(frame.Fields))
			for idx, value := range frame.Fields {
				if value.Name == "time" {
					vals[0] = time.Now()
				} else {
					vals[idx] = dataItem[value.Name]
					idx++
				}
			}
			frame.AppendRow(vals...)
		}
	} else {
		backend.Logger.Error("QueryingResponse is nil")
	}
	return frame
}
