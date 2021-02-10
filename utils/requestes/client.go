package requestes

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type RequestsConfig struct {
	VerifyTLS  bool
	CaCertPath string
	MutualTLS  bool
	CaKeyPath  string
	CaCrtPath  string
}

type RequestsClient struct {
	client *http.Client
}

type ExtraConfig func(r *http.Request)

type PostData func(url string) (*http.Request, error)

type ResponseData struct {
	status string
	statusCode int
	header map[string][]string
	data []byte
	//data string
}

func (rsp ResponseData) Text() string {
	return string(rsp.data[:])
}

func (rsp ResponseData) BindJSON(obj interface{}) error {
	if err := json.Unmarshal(rsp.data, obj); err != nil {
		return err
	}
	return nil
}

func New(c RequestsConfig) (*RequestsClient, error) {

	pool := x509.NewCertPool()
	if c.CaCertPath != "" {
		caCrt, err := ioutil.ReadFile(c.CaCertPath)
		if err != nil {
			return &RequestsClient{}, errors.New(fmt.Sprintf("read ca cert file error: %s", err))
		}
		pool.AppendCertsFromPEM(caCrt)
	}

	Certificates := make([]tls.Certificate, 0)
	if c.MutualTLS {
		if c.CaKeyPath == "" || c.CaCrtPath == "" {
			return &RequestsClient{}, errors.New("config error: if use mutual TLS, Must specify both of caKeyPath and caCrtPath")
		}
		cliCrt, err := tls.LoadX509KeyPair(c.CaCrtPath, c.CaKeyPath)
		if err != nil {
			return &RequestsClient{}, errors.New(fmt.Sprintf("load x509 key pair error: %s", err))
		}
	    Certificates = []tls.Certificate{cliCrt}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !c.VerifyTLS,
			RootCAs:      pool,
			Certificates: Certificates,
		},
	}

	return &RequestsClient{&http.Client{Transport: tr}}, nil
}

func AddHeader(data map[string]string) ExtraConfig {
	add := func(req *http.Request) {
		for k, v := range data {
			req.Header.Set(k, v)
		}
	}
	return add
}

//func AddQueryParam(data map[string]string) ExtraConfig {
//	add := func(req *http.Request) {
//		q := req.URL.Query()
//		for k, v := range data {
//			q.Set(k, v)
//		}
//		req.URL.RawQuery = q.Encode()
//	}
//	return add
//}

// TODO: Ugly, need format trans, only support simple map/struct, optimize it
func AddQueryParam(data interface{}) ExtraConfig {
	dataType := reflect.TypeOf(data).Kind().String()
	if dataType == "map" {
		add := func(req *http.Request) {
			q := req.URL.Query()
			for k, v := range data.(map[string]string) {
				q.Set(k, v)
			}
			req.URL.RawQuery = q.Encode()
		}
		return add
	} else if dataType == "struct" {
		add := func(req *http.Request) {
			q := req.URL.Query()
			rType := reflect.TypeOf(data)
			rVal := reflect.ValueOf(data)
			for k := 0; k < rVal.NumField(); k++ {
				q.Set(rType.Field(k).Name, rVal.Field(k).String())
			}
			req.URL.RawQuery = q.Encode()
		}
		return add
	} else {
		fmt.Println("AddQueryParam now can only support a map or struct")
		return func(req *http.Request) {}
	}
}

func FormData(data map[string]string) PostData {
    add := func(url string) (*http.Request, error) {

		list := make([]string, 0)
		for k, v := range data {
			list = append(list, fmt.Sprintf("%s=%s", k, fmt.Sprint(v)))
		}
		data :=strings.Join(list, "&")

		req, err := http.NewRequest("POST", url, strings.NewReader(data))
		if err != nil {
			return &http.Request{}, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		return req, nil

	}

	return add
}

func JsonData(data interface{}) PostData {
	add := func(url string) (*http.Request, error) {
		b ,err := json.Marshal(data)
		if err != nil {
			return &http.Request{}, errors.New(fmt.Sprintf("json format error: %s", err))
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
		if err != nil {
			return &http.Request{}, errors.New(fmt.Sprintf("create request error: %s", err))
		}
		req.Header.Add("Content-Type", "application/json;charset=utf-8")

		return req, nil
	}

	return add
}

func generateRepData(resp *http.Response) (ResponseData, error) {
	status := resp.Status
	statusCode := resp.StatusCode
	headers := make(map[string][]string)

	for k, v := range resp.Header {
		headers[k] = v
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ResponseData{status: status, statusCode: statusCode, header: headers, data: nil}, errors.New(fmt.Sprintf("read response data error cause: %s", err))
	}

	return ResponseData{status: status, statusCode: statusCode, header: headers, data: body}, nil
}

func (r *RequestsClient) Get(url string, extraConfigs ...ExtraConfig) (ResponseData, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return ResponseData{}, errors.New(fmt.Sprintf("post method create request error cause: %s", err))
	}

	for _, extraConfig := range extraConfigs {
		extraConfig(req)
	}

	resp, err := r.client.Do(req)

	if err != nil {
		return ResponseData{}, errors.New(fmt.Sprintf("do get error cause: %s", err))
	}

	return generateRepData(resp)
}

func (r *RequestsClient) Post(url string, postData PostData, extraConfigs ...ExtraConfig) (ResponseData, error) {
	req, err := postData(url)
	if err != nil {
		return ResponseData{}, errors.New(fmt.Sprintf("post method get error: %s", err))
	}
	for _, extraConfig := range extraConfigs {
		extraConfig(req)
	}

	resp, err := r.client.Do(req)

	if err != nil {
		return ResponseData{}, errors.New(fmt.Sprintf("do post error cause: %s", err))
	}

	return generateRepData(resp)
}