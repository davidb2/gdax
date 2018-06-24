package gdax

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/imdario/mergo"
	"net/http"
	"time"
)

const (
	Fills = "fills"
	Pdf   = "pdf"
	Csv   = "csv"
)

type ReportParams struct {
	StartDate *time.Time `json:"start_date,string,omitempty"`
	EndDate   *time.Time `json:"end_date,string,omitempty"`
}

type Report struct {
	Type      string     `json:"type"`
	StartDate *time.Time `json:"start_date,string"`
	EndDate   *time.Time `json:"end_date,string"`
	ProductId string     `json:"product_id,omitempty"`
	AccountId *uuid.UUID `json:"account_id,string,omitempty"`
	Format    string     `json:"format,omitempty"`
	Email     string     `json:"email,omitempty"`

	// response params
	Id          *uuid.UUID    `json:"id,string,omitempty"`
	Status      string        `json:"status,omitempty"`
	CreatedAt   *time.Time    `json:"created_at,string,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,string,omitempty"`
	ExpiresAt   *time.Time    `json:"expires_at,string,omitempty"`
	FileUrl     string        `json:"file_url,omitempty"`
	Params      *ReportParams `json:"params,omitempty"`
}

func (accessInfo *AccessInfo) CreateReport(report *Report) (*Report, error) {
	// POST /reports
	var reportResponse Report
	jsonBytes, err := json.Marshal(*report)
	if err != nil {
		return nil, err
	}
	_, err = accessInfo.request(http.MethodPost, "/reports", string(jsonBytes), &reportResponse)
	if err != nil {
		return nil, err
	}
	if err = mergo.Merge(&reportResponse, *report); err != nil {
		return nil, err
	}

	return &reportResponse, err
}

func (accessInfo *AccessInfo) GetReportStatus(reportId *uuid.UUID) (*Report, error) {
	// GET /reports/:report_id
	var reportStatus Report
	_, err := accessInfo.request(http.MethodGet, "/reports/"+reportId.String(), "", &reportStatus)
	if err != nil {
		return nil, err
	}
	return &reportStatus, nil
}
