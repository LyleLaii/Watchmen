package promclient

import "time"

type QueryParam struct {
	query string
	time string //<rfc3339 | unix_timestamp>
	timeout string //<duration>
}

type RangerQueryParam struct {
	query string
	start string //<rfc3339 | unix_timestamp>
	end string //<rfc3339 | unix_timestamp>
	step string //<duration | float>
	timeout string //<duration>
}

type QueryData struct {
	Label  map[string]string `json:"metric"`
	Metric []interface{}     `json:"value"`
}

type RangerData struct {
	Label   map[string]string `json:"metric"`
	Metrics [][]interface{}   `json:"values"`
}

type QueryResult struct {
	Status string	`json:"status"`
	Data 	struct{ResultType string `json:"resultType"`
	                Result []QueryData `json:"result"`}  `json:"data"`
}

type RangerQueryResult struct {
	Status string	`json:"status"`
	Data 	struct{ResultType string `json:"resultType"`
		Result []RangerData `json:"result"`}  `json:"data"`
}

// prometheus return value [float64 string]
func ParsPromTime(timestamp interface{}) time.Time {
	ts := timestamp.(float64)
	secs := int64(ts)
	nsecs := int64((ts - float64(secs)) * 1e9)
	t := time.Unix(secs, nsecs)
	return t
}