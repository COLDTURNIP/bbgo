// Code generated by "requestgen -method GET -type MiningRequest -url /v1/metrics/mining/:metric -responseType Response"; DO NOT EDIT.

package glassnode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

func (m *MiningRequest) SetAsset(Asset string) *MiningRequest {
	m.Asset = Asset
	return m
}

func (m *MiningRequest) SetSince(Since int64) *MiningRequest {
	m.Since = Since
	return m
}

func (m *MiningRequest) SetUntil(Until int64) *MiningRequest {
	m.Until = Until
	return m
}

func (m *MiningRequest) SetInterval(Interval Interval) *MiningRequest {
	m.Interval = Interval
	return m
}

func (m *MiningRequest) SetFormat(Format Format) *MiningRequest {
	m.Format = Format
	return m
}

func (m *MiningRequest) SetTimestampFormat(TimestampFormat string) *MiningRequest {
	m.TimestampFormat = TimestampFormat
	return m
}

func (m *MiningRequest) SetMetric(Metric string) *MiningRequest {
	m.Metric = Metric
	return m
}

// GetQueryParameters builds and checks the query parameters and returns url.Values
func (m *MiningRequest) GetQueryParameters() (url.Values, error) {
	var params = map[string]interface{}{}
	// check Asset field -> json key a
	Asset := m.Asset

	// TEMPLATE check-required
	if len(Asset) == 0 {
		return nil, fmt.Errorf("a is required, empty string given")
	}
	// END TEMPLATE check-required

	// assign parameter of Asset
	params["a"] = Asset
	// check Since field -> json key s
	Since := m.Since

	// assign parameter of Since
	params["s"] = Since
	// check Until field -> json key u
	Until := m.Until

	// assign parameter of Until
	params["u"] = Until
	// check Interval field -> json key i
	Interval := m.Interval

	// assign parameter of Interval
	params["i"] = Interval
	// check Format field -> json key f
	Format := m.Format

	// assign parameter of Format
	params["f"] = Format
	// check TimestampFormat field -> json key timestamp_format
	TimestampFormat := m.TimestampFormat

	// assign parameter of TimestampFormat
	params["timestamp_format"] = TimestampFormat

	query := url.Values{}
	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParameters builds and checks the parameters and return the result in a map object
func (m *MiningRequest) GetParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}

	return params, nil
}

// GetParametersQuery converts the parameters from GetParameters into the url.Values format
func (m *MiningRequest) GetParametersQuery() (url.Values, error) {
	query := url.Values{}

	params, err := m.GetParameters()
	if err != nil {
		return query, err
	}

	for k, v := range params {
		query.Add(k, fmt.Sprintf("%v", v))
	}

	return query, nil
}

// GetParametersJSON converts the parameters from GetParameters into the JSON format
func (m *MiningRequest) GetParametersJSON() ([]byte, error) {
	params, err := m.GetParameters()
	if err != nil {
		return nil, err
	}

	return json.Marshal(params)
}

// GetSlugParameters builds and checks the slug parameters and return the result in a map object
func (m *MiningRequest) GetSlugParameters() (map[string]interface{}, error) {
	var params = map[string]interface{}{}
	// check Metric field -> json key metric
	Metric := m.Metric

	// assign parameter of Metric
	params["metric"] = Metric

	return params, nil
}

func (m *MiningRequest) applySlugsToUrl(url string, slugs map[string]string) string {
	for k, v := range slugs {
		needleRE := regexp.MustCompile(":" + k + "\\b")
		url = needleRE.ReplaceAllString(url, v)
	}

	return url
}

func (m *MiningRequest) GetSlugsMap() (map[string]string, error) {
	slugs := map[string]string{}
	params, err := m.GetSlugParameters()
	if err != nil {
		return slugs, nil
	}

	for k, v := range params {
		slugs[k] = fmt.Sprintf("%v", v)
	}

	return slugs, nil
}

func (m *MiningRequest) Do(ctx context.Context) (Response, error) {

	// no body params
	var params interface{}
	query, err := m.GetQueryParameters()
	if err != nil {
		return nil, err
	}

	apiURL := "/v1/metrics/mining/:metric"
	slugs, err := m.GetSlugsMap()
	if err != nil {
		return nil, err
	}

	apiURL = m.applySlugsToUrl(apiURL, slugs)

	req, err := m.Client.NewAuthenticatedRequest(ctx, "GET", apiURL, query, params)
	if err != nil {
		return nil, err
	}

	response, err := m.Client.SendRequest(req)
	if err != nil {
		return nil, err
	}

	var apiResponse Response
	if err := response.DecodeJSON(&apiResponse); err != nil {
		return nil, err
	}
	return apiResponse, nil
}