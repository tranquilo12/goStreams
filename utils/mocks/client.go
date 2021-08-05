package mocks

import (
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
