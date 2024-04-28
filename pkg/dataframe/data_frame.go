package dataframe

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/pavi6691/go-boilingdata/models"
)

func initiateNewDataFrame(refID string, response *models.Response) *data.Frame {
	firstDataEntry := response.Data[0]
	frame := data.NewFrame("response")
	// add fields
	frame.RefID = refID
	var field *data.Field
	for indx, key := range response.Keys {
		var ok bool = false
		if indx == 0 {
			_, err := parseDateTime(firstDataEntry[key])
			if err == nil {
				ok = true
			}
		}
		if ok {
			field = data.NewField("time", nil, []time.Time{})
		} else {
			switch firstDataEntry[key].(type) {
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
				panic(fmt.Errorf("ksy '%s' value '%s' specified with unsupported type", key, firstDataEntry[key]))
			}
		}
		frame.Fields = append(frame.Fields, field)
	}
	return frame
}

func getValue(value interface{}) string {
	switch v := value.(type) {
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.UTC().String()
	case json.RawMessage:
		return string(v)
	case data.EnumItemIndex:
		return string(v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func parseDateTime(dateDate interface{}) (time.Time, error) {
	switch v := dateDate.(type) {
	case time.Time:
		return v, nil // If value is already a time.Time, return it
	case string:
		return time.Parse(time.RFC3339Nano, v) // If value is a string, parse it as time
	case int64:
		return time.UnixMilli(v), nil // If value is an int64, convert it to time.Time
	case float64:
		return time.UnixMilli(int64(v)), nil // If value is an int64, convert it to time.Time
	default:
		return time.Time{}, fmt.Errorf("unsupported type %T", dateDate)
	}
}

func GetFrames(refID string, response *models.Response) (*data.Frame, error) {
	var frame *data.Frame
	if response != nil && len(response.Data) > 0 {
		frame = initiateNewDataFrame(refID, response)
		for _, dataItem := range response.Data {
			vals := make([]interface{}, len(frame.Fields))
			for idx, value := range frame.Fields {
				if value.Name == "time" {
					validTime, err := parseDateTime(dataItem[response.Keys[0]])
					if err == nil {
						vals[0] = validTime
					} else {
						v := getValue(dataItem[response.Keys[0]])
						return nil, fmt.Errorf("This value = %s in Column %s cannot be converted to date and time", v, response.Keys[0])
					}
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
	return frame, nil
}
