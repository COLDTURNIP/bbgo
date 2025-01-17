package binanceapi

import (
	"encoding/json"
	"time"

	"github.com/c9s/requestgen"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

// rebate type：1 is commission rebate，2 is referral kickback
type RebateType int

const (
	RebateTypeCommission       = 1
	RebateTypeReferralKickback = 2
)

type SpotRebate struct {
	Asset      string                     `json:"asset"`
	Type       RebateType                 `json:"type"`
	Amount     fixedpoint.Value           `json:"amount"`
	UpdateTime types.MillisecondTimestamp `json:"updateTime"`
}

type PagedResponse struct {
	Status string `json:"status"`
	Type   string `json:"type"`
	Code   string `json:"code"`
	Data   struct {
		Page         int             `json:"page"`
		TotalRecords int             `json:"totalRecords"`
		TotalPageNum int             `json:"totalPageNum"`
		Data         json.RawMessage `json:"data"`
	} `json:"data"`
}

// GetSpotRebateHistoryRequest
// The max interval between startTime and endTime is 30 days.
// If startTime and endTime are not sent, the recent 7 days' data will be returned.
// The earliest startTime is supported on June 10, 2020
//go:generate requestgen -method GET -url "/sapi/v1/rebate/taxQuery" -type GetSpotRebateHistoryRequest -responseType PagedResponse -responseDataField Data.Data -responseDataType []SpotRebate
type GetSpotRebateHistoryRequest struct {
	client requestgen.AuthenticatedAPIClient

	startTime *time.Time `param:"startTime,milliseconds"`
	endTime   *time.Time `param:"endTime,milliseconds"`
}

func (c *RestClient) NewGetSpotRebateHistoryRequest() *GetSpotRebateHistoryRequest {
	return &GetSpotRebateHistoryRequest{client: c}
}
