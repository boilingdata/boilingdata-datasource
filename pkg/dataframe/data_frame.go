package dataframe

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boilingdata/go-boilingdata/messages"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

func initiateNewDataFrame(refID string, response *messages.Response) *data.Frame {
	firstDataEntry := response.Data[0]
	frame := data.NewFrame("response")
	// add fields
	frame.RefID = refID
	var field *data.Field
	for indx, key := range response.Keys {
		var ok bool = false
		if indx == 0 {
			times, err := parseDateTime(firstDataEntry[key])
			if err == nil && times.Unix() > 0 {
				ok = true
			}
		}
		if ok {
			field = data.NewFieldFromFieldType(data.FieldTypeNullableTime, len(response.Data))
			field.Name = "time"
		} else {
			switch firstDataEntry[key].(type) {
			case int8:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableInt8, len(response.Data))
			case int16:
				field = data.NewFieldFromFieldType(data.FieldTypeInt16, len(response.Data))
			case int32:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableInt32, len(response.Data))
			case int64:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableInt64, len(response.Data))
			case uint8:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableUint8, len(response.Data))
			case uint16:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableUint16, len(response.Data))
			case uint32:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableUint32, len(response.Data))
			case uint64:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableUint64, len(response.Data))
			case float32:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableFloat32, len(response.Data))
			case float64:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableFloat64, len(response.Data))
			case string:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableString, len(response.Data))
			case bool:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableBool, len(response.Data))
			case time.Time:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableTime, len(response.Data))
			case json.RawMessage:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableJSON, len(response.Data))
			case data.EnumItemIndex:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableEnum, len(response.Data))
			default:
				field = data.NewFieldFromFieldType(data.FieldTypeNullableString, len(response.Data))
			}
			field.Name = key
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

func parseDateTime(dateTime interface{}) (time.Time, error) {
	switch v := dateTime.(type) {
	case time.Time:
		return v, nil // If value is already a time.Time, return it
	case string:
		formats := []string{
			time.RFC3339Nano,
			"2006-01-02 15:04:05+00",
			time.DateTime,
			time.Layout,
			time.ANSIC,
			time.UnixDate,
			time.RubyDate,
			time.RFC822,
			time.RFC822Z,
			time.RFC850,
			time.RFC1123,
			time.RFC1123Z,
			time.RFC3339,
			time.Kitchen,
			time.Stamp,
			time.StampMilli,
			time.StampMicro,
			time.StampNano,
			time.DateOnly,
			time.TimeOnly,
		}
		for _, format := range formats {
			t, err := time.Parse(format, v)
			if err == nil {
				return t, nil
			}
		}
		return time.Time{}, fmt.Errorf("failed to parse time from string %q", v)
	case int64:
		return time.Unix(0, v*int64(time.Millisecond)), nil // If value is an int64, convert it to time.Time
	case float64:
		return time.Unix(0, int64(v)*int64(time.Millisecond)), nil // If value is an int64, convert it to time.Time
	default:
		return time.Time{}, fmt.Errorf("unsupported type %T", dateTime)
	}
}

func GetFrames(refID string, response *messages.Response) (*data.Frame, error) {
	var frame *data.Frame
	if response != nil && len(response.Data) > 0 {
		frame = initiateNewDataFrame(refID, response)
		for idx, dataItem := range response.Data {
			for keyIndx, key := range response.Keys {
				field := frame.Fields[keyIndx]
				if dataItem[key] != nil {
					if field.Name == "time" {
						validTime, err := parseDateTime(dataItem[key])
						if err == nil {
							field.SetConcrete(idx, validTime)
						} else {
							v := getValue(dataItem[key])
							return nil, fmt.Errorf("This value = %s in Column %s cannot be converted to date and time", v, response.Keys[0])
						}
					} else {
						field.SetConcrete(idx, dataItem[key])
					}
				}
			}
		}
	} else {
		backend.Logger.Error("No response")
		return nil, fmt.Errorf("No response")
	}
	return frame, nil
}
