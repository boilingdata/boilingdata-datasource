package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/boilingdata/boilingdata/pkg/dataframe"
	"github.com/boilingdata/boilingdata/pkg/settings"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/pavi6691/go-boilingdata/boilingdata"
	"github.com/pavi6691/go-boilingdata/models"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &Datasource{}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type QueryModel struct {
	SelectQuery string `json:"selectQuery"`
	UUID        string `json:"uuid"`
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse
	// Unmarshal the JSON into our queryModel.
	var qm QueryModel
	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		backend.Logger.Error("error unmarshalling QueryModel : " + err.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("error unmarshalling QueryModel: %v", err.Error()))
	}
	payload := models.GetPayLoad()
	payload.RequestID = qm.UUID + "-" + query.RefID
	payload.SQL = qm.SelectQuery
	// Convert the payload to JSON string
	jsonQuery, err := json.Marshal(payload)
	if err != nil {
		backend.Logger.Error("error marshalling jsonQuery : " + err.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("error marshalling: %v", err.Error()))
	}
	config, err := settings.LoadPluginSettings(*pCtx.DataSourceInstanceSettings)
	if err != nil {
		backend.Logger.Error("Unable to load settings : " + err.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("Unable to load settings : %v", err.Error()))
	}
	userService := boilingdata.GetInstance(config.UserName, config.Secrets.Password)
	queryResponse, err := userService.Query(jsonQuery)
	if err != nil {
		backend.Logger.Error("Error while querying : " + err.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("Error while querying : %v", err.Error()))
	}
	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames
	frame, err := dataframe.GetFrames(query.RefID, queryResponse)
	if err != nil {
		backend.Logger.Error(err.Error())
		return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
	}
	// add the frames to the response.
	response.Frames = append(response.Frames, frame)
	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, err := settings.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	if config.Secrets.Password == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Please enter password"
		return res, nil
	}

	if config.UserName == "" {
		res.Status = backend.HealthStatusError
		res.Message = "Please enter username"
		return res, nil
	}
	instance := boilingdata.GetInstance(config.UserName, config.Secrets.Password)
	_, err = instance.Auth.Authenticate()

	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Authentication failed"
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Authentication successfull",
	}, nil
}
