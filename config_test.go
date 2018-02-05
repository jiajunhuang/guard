package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	badJSON    = `"name":"www.example.com","backends":["127.0.0.1:80","127.0.0.1:80","127.0.0.1:80"],"weights":[5,1],"ratio":0.3,"paths":["/"],"methods":["GET"]}`
	badConfig  = `{"name":"www.example.com","backends":["127.0.0.1:80","127.0.0.1:80","127.0.0.1:80"],"weights":[5,1],"ratio":0.3,"paths":["/"],"methods":["GET"]}`
	goodConfig = `{"name":"www.example.com","backends":["127.0.0.1:80","127.0.0.1:80","127.0.0.1:80"],"weights":[5,1,1],"ratio":0.3,"paths":["/"],"methods":["GET"]}`
)

func TestUpdateConfig(t *testing.T) {
	fakeServer := httptest.NewServer(
		http.HandlerFunc(appHandler),
	)
	defer fakeServer.Close()

	url := fakeServer.URL + "/app"

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

	// check fallback
	config.Name = "www.example.com"
	config.Backends = []string{"192.168.1.1:80"}
	config.Weights = []int{1}
	config.Ratio = 0.3
	config.DisableTSR = false
	config.LoadBalanceMethod = LBMWRR
	config.Paths = []string{"/"}
	config.Methods = []string{"get"}

	// "" or text
	config.FallbackType = ""
	content := "too many requests"
	if err := checkAppConfig(config); err != nil {
		t.Errorf("should not return error but got: %s", err)
	}
	if config.FallbackType != fallbackTEXT || config.FallbackContent != content {
		t.Errorf("fallback options error: %s, %s", config.FallbackType, config.FallbackContent)
	}
	config.FallbackType = fallbackTEXT
	if err := checkAppConfig(config); err != nil {
		t.Errorf("should not return error but got: %s", err)
	}
	if config.FallbackType != fallbackTEXT || config.FallbackContent != content {
		t.Errorf("fallback options error: %s, %s", config.FallbackType, config.FallbackContent)
	}

	// "json"
	config.FallbackType = fallbackJSON
	content = "{}"
	config.FallbackContent = content
	if err := checkAppConfig(config); err != nil {
		t.Errorf("should not return error but got: %s", err)
	}
	if config.FallbackType != fallbackJSON || config.FallbackContent != content {
		t.Errorf("fallback options error: %s, %s", config.FallbackType, config.FallbackContent)
	}

	// "html"
	config.FallbackType = fallbackHTML
	content = "<html></html>"
	config.FallbackContent = content
	if err := checkAppConfig(config); err != nil {
		t.Errorf("should not return error but got: %s", err)
	}
	if config.FallbackType != fallbackHTML || config.FallbackContent != content {
		t.Errorf("fallback options error: %s, %s", config.FallbackType, config.FallbackContent)
	}

	// "html_file"
	// file not exist
	filePath := "/tmp/test_guard_html_file_path.html"
	os.Remove(filePath)
	config.FallbackType = fallbackHTMLFile
	config.FallbackContent = filePath
	if err := checkAppConfig(config); err == nil {
		t.Errorf("should return error but not")
	}

	// file exists
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("failed to open %s: %s", filePath, err)
	}
	defer f.Close()
	content = "<html></html>"
	f.Write([]byte(content))
	if err := checkAppConfig(config); err != nil {
		t.Errorf("should not return error, but got: %s", err)
	}
	if config.FallbackType != fallbackHTML || config.FallbackContent != content {
		t.Errorf("fallback options error: %s, %s", config.FallbackType, config.FallbackContent)
	}

	// bad type
	config.FallbackType = "what"
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

	readContent, _ := readFromFile(*configPath)
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
		"",
		"too many requests",
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
	f.Truncate(0)
	f.Seek(0, 0)
	f.Write([]byte(goodConfig))
	f.Close()

	go configKeeper()

	close(configSync)
}

func TestConfigIndex(t *testing.T) {
	fakeServer := httptest.NewServer(
		http.HandlerFunc(configIndexHandler),
	)
	defer fakeServer.Close()
	url := fakeServer.URL + "/"

	// test 200
	f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Errorf("failed to open config file: %s", err)
	}
	f.Truncate(0)
	f.Seek(0, 0)
	f.Write([]byte(goodConfig))
	f.Close()

	resp, err := http.Get(url)
	if err != nil {
		t.Errorf("failed to get config index page")
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("should got 200, but: %d", resp.StatusCode)
	}
}
