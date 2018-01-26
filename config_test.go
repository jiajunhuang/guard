package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestReadFromFile(t *testing.T) {
	os.Remove(*configPath)
	defer os.Remove(*configPath)

	f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("failed to open config file: %s", err)
	}
	defer f.Close()
	content := []byte("hello world")
	f.Write(content)

	readContent := readFromFile(*configPath)
	if !bytes.Equal(readContent, content) {
		t.Errorf("read from file return bad content: %s", string(readContent))
	}
}

func TestConfigKeeper(t *testing.T) {
	os.Remove(*configPath)
	defer os.Remove(*configPath)

	config := appConfig{
		"www.example.com",
		[]string{"192.168.1.1:80"},
		[]int{1},
		0.3,
		false,
		LBMWRR,
		[]string{"/"},
		[]string{"GET"},
	}

	go configKeeper()

	configSync <- config
}

func TestConfigKeeperBadJSON(t *testing.T) {
	os.Remove(*configPath)
	defer os.Remove(*configPath)
	f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("failed to open config file: %s", err)
	}
	defer f.Close()
	content := []byte("hello world")
	f.Write(content)

	go configKeeper()
}

func TestConfigKeeperGoodJSON(t *testing.T) {
	os.Remove(*configPath)
	//defer os.Remove(*configPath)
	f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("failed to open config file: %s", err)
	}
	config := appConfig{
		"www.example.com",
		[]string{"192.168.1.1:80"},
		[]int{1},
		0.3,
		false,
		LBMWRR,
		[]string{"/"},
		[]string{"GET"},
	}
	breakerConfig := make(map[string]appConfig)
	breakerConfig[config.Name] = config
	fileBytes, err := json.Marshal(breakerConfig)
	if err != nil {
		t.Errorf("failed to marshal breaker config: %s", err)
	}
	f.Truncate(0)
	f.Seek(0, 0)
	f.Write(fileBytes)
	f.Close()

	go configKeeper()

	close(configSync)
}
