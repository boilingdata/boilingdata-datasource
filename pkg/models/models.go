package models

type QueryModel struct {
	SelectQuery string `json:"selectQuery"`
}
type payload struct {
	MessageType string `json:"messageType"`
	SQL         string `json:"sql"`
	RequestID   string `json:"requestId"`
	ReadCache   string `json:"readCache"`
	Tags        []Tag  `json:"tags"`
}

func GetPayLoad() payload {
	return payload{
		MessageType: "SQL_QUERY",
		SQL:         "",
		RequestID:   "",
		ReadCache:   "NONE",
		Tags: []Tag{
			{
				Name:  "CostCenter",
				Value: "930",
			},
			{
				Name:  "ProjectId",
				Value: "Top secret Area 53",
			},
		},
	}
}

// Define structs to represent the JSON payload
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
