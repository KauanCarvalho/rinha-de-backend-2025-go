package processor

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

var fallbackBaseURL = getEnv("FALLBACK_PROCESSOR_URL", "http://payment-processor-fallback:8080")

var fallbackClient = &fasthttp.Client{
	MaxConnsPerHost:           4096,
	MaxIdleConnDuration:       90 * time.Second,
	ReadTimeout:               5 * time.Second,
	WriteTimeout:              5 * time.Second,
	DisableHeaderNamesNormalizing: true,
}

func FallbackCreatePayment(req PaymentRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)
	defer fasthttp.ReleaseResponse(response)

	request.SetRequestURI(fallbackBaseURL + "/payments")
	request.Header.SetMethod(fasthttp.MethodPost)
	request.Header.SetContentType("application/json")
	request.SetBody(body)

	if err := fallbackClient.Do(request, response); err != nil {
		return err
	}

	if response.StatusCode() != fasthttp.StatusOK {
		return errors.New("unexpected status code on payment request")
	}

	return nil
}

func FallbackHealthcheck() (*HealthResponse, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fallbackBaseURL + "/payments/service-health")
	req.Header.SetMethod(fasthttp.MethodGet)

	if err := fallbackClient.Do(req, resp); err != nil {
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
