package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateConfig(t *testing.T) {
	fakeServer := httptest.NewServer(
		http.HandlerFunc(appHandler),
	)
	defer fakeServer.Close()

	url := fakeServer.URL + "/app"
	badJSON := `"name":"www.example.com","backends":["127.0.0.1:80","127.0.0.1:80","127.0.0.1:80"],"weights":[5,1],"ratio":0.3,"paths":["/"],"methods":["GET"]}`
	badConfig := `{"name":"www.example.com","backends":["127.0.0.1:80","127.0.0.1:80","127.0.0.1:80"],"weights":[5,1],"ratio":0.3,"paths":["/"],"methods":["GET"]}`
	goodConfig := `{"name":"www.example.com","backends":["127.0.0.1:80","127.0.0.1:80","127.0.0.1:80"],"weights":[5,1,1],"ratio":0.3,"paths":["/"],"methods":["GET"]}`

	// try get
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("should return 405, but got: %d", resp.StatusCode)
	}

	// post bad json
	resp, err = http.Post(url, "application/json", bytes.NewBuffer([]byte(badJSON)))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Errorf("should return 400, but got: %d", resp.StatusCode)
	}

	// post bad config
	resp, err = http.Post(url, "application/json", bytes.NewBuffer([]byte(badConfig)))
	if err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Errorf("should return 400, but got: %d", resp.StatusCode)
	}

	// post bad config
	resp, err = http.Post(url, "application/json", bytes.NewBuffer([]byte(goodConfig)))
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Errorf("should return 200, but got: %d", resp.StatusCode)
	}
}

func TestCheckAPPConfig(t *testing.T) {
	config := &appConfig{}

	// not not exist
	if err := checkAppConfig(config); err == nil {
		t.Errorf("should return error but not")
	}

	config.Name = "www.example.com"
	// len(backends) != len(weights)
	config.Backends = []string{"192.168.1.1:80"}
	if err := checkAppConfig(config); err == nil {
		t.Errorf("should return error but not")
	}

	// len(path) != len(methods)
	config.Weights = []int{1}
	config.Paths = []string{"/"}
	if err := checkAppConfig(config); err == nil {
		t.Errorf("should return error but not")
	}
	config.Methods = []string{"get"}

	// load balancer algorithm not exist
	config.LoadBalanceMethod = "what"
	if err := checkAppConfig(config); err == nil {
		t.Errorf("should return error but not")
	}
}

func TestGetBalancer(t *testing.T) {
	getBalancer(LBMWRR)
	getBalancer(LBMRR)
	getBalancer(LBMRandom)

	defer shouldPanic()
	getBalancer("what")
}
