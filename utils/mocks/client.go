package mocks

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

// MockClient Define a client that contains a modifiable 'DoFunc'
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// GetDoFunc define a function that will take a request and return a response
var (
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// Do Overwrite the 'Do' function of Mock Client
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

// MockHTTPResponse make function that mocks the HTTPResponse
func MockHTTPResponse(body string, status int) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(body)))
	response := &http.Response{StatusCode: status, Body: r}
	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return response, nil
	}
}

// MockHTTPError make function to mock errors
func MockHTTPError(err string) {
	GetDoFunc = func(r *http.Request) (*http.Response, error) {
		return nil, errors.New(err)
	}
}
