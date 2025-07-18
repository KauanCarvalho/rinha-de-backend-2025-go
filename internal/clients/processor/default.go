package processor

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

var defaultBaseURL = getEnv("DEFAULT_PROCESSOR_URL", "http://payment-processor-default:8080")

var defaultClient = &fasthttp.Client{
	MaxConnsPerHost:               8192,
	MaxIdleConnDuration:           90 * time.Second,
	ReadTimeout:                   5 * time.Second,
	WriteTimeout:                  5 * time.Second,
	DisableHeaderNamesNormalizing: true,
}

func DefaultCreatePayment(req PaymentRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.SetRequestURI(defaultBaseURL + "/payments")
	request.Header.SetMethod(fasthttp.MethodPost)
	request.Header.SetContentType("application/json")
	request.SetBody(body)

	if err := defaultClient.Do(request, response); err != nil {
		return err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		return errors.New("unexpected status code on payment request")
	}

	return nil
}

func DefaultHealthcheck() (*HealthResponse, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(defaultBaseURL + "/payments/service-health")
	req.Header.SetMethod(fasthttp.MethodGet)

	if err := defaultClient.Do(req, resp); err != nil {
		return nil, err
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		return nil, errors.New("unexpected status code")
	}

	var result HealthResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
