package promclient

import (
	"github.com/pkg/errors"
	"time"
	"watchmen/utils/requestes"
)

const (
	InstanceQuery = "/api/v1/query"
	RangeQuery = "/api/v1/query_range"
)

type PromClient struct {
	promURL string
	httpClient *requestes.RequestsClient
}

type PromConfig struct {
	Url        string `json:"url,omitempty"`
	VerifyTLS  bool `json:"verify_tls,omitempty"`
	CaCertPath string `json:"ca_cert_path,omitempty"`
	MutualTLS  bool `json:"mutual_tls,omitempty"`
	CaKeyPath  string `json:"ca_key_path,omitempty"`
	CaCrtPath  string `json:"ca_crt_path,omitempty"`

}

func New(conf PromConfig) (*PromClient, error) {
	hc, err := requestes.New(requestes.RequestsConfig{VerifyTLS: conf.VerifyTLS,
													CaCertPath:conf.CaCrtPath,
													MutualTLS:conf.MutualTLS,
													CaKeyPath:conf.CaKeyPath,
													CaCrtPath:conf.CaCrtPath})
	if err != nil {
		return &PromClient{
			promURL:    "",
			httpClient: nil,
		}, errors.Wrap(err, "create PromClient failed")
	}

	return &PromClient{
		promURL:    conf.Url,
		httpClient: hc,
	}, nil

}

func (p *PromClient) Query(param QueryParam) (QueryResult, error) {
	data, err := p.httpClient.Get(p.promURL + InstanceQuery, requestes.AddQueryParam(param))
	if err != nil {
		return QueryResult{}, errors.Wrap(err, "query data failed")
	}
	d := QueryResult{}

	if err := data.BindJSON(&d); err != nil {
		return QueryResult{}, errors.Wrap(err, "query data failed")
	}
	return d, nil
}

func (p *PromClient) RangeQuery(param RangerQueryParam) (RangerQueryResult, error) {
	data, err := p.httpClient.Get(p.promURL + RangeQuery, requestes.AddQueryParam(param))
	if err != nil {
		//fmt.Printf("Query Instance Data error: %v", err)
		return RangerQueryResult{}, errors.Wrap(err, "query instance data failed")
	}
	d := RangerQueryResult{}

	if err := data.BindJSON(&d); err != nil {
		return RangerQueryResult{}, errors.Wrap(err, "query data failed")
	}

	return d, nil
}


func (p *PromClient) FetchRangeData(queryStr string, dayRange time.Duration, timeStep string) (RangerQueryResult, error) {
	now := time.Now()
	pastDate := now.Add(-dayRange)
	nowStr := now.Format(time.RFC3339)
	pastDateStr := pastDate.Format(time.RFC3339)
	rqp := RangerQueryParam{
		query:   queryStr,
		start:   pastDateStr,
		end:     nowStr,
		step:    timeStep,
		timeout: "",
	}

	return p.RangeQuery(rqp)
}
