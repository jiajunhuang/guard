package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

var (
	errNameEmpty               = errors.New("name is required")
	errBackendWeightNotMatch   = errors.New("backend and weight does not match")
	errPathMethodNotMatch      = errors.New("path and method does not match")
	errBadLoadBalanceAlgorithm = errors.New("bad load balance algorithm, only wrr, rr, random are support now")
)

type appConfig struct {
	Name              string   `json:"name"`
	Backends          []string `json:"backends"` // e.g. ["192.168.1.1:80", "192.168.1.2:80", "192.168.1.3:1080"]
	Weights           []int    `json:"weights"`  // e.g. [5, 1, 1]
	Ratio             float64  `json:"ratio"`
	DisableTSR        bool     `json:"disable_tsr"`
	LoadBalanceMethod string   `json:"load_balance_method"` // wrr, rr, random
	Paths             []string `json:"paths"`
	Methods           []string `json:"methods"`
}

func checkAppConfig(a *appConfig) error {
	if a.Name == "" {
		return errNameEmpty
	}

	if len(a.Backends) != len(a.Weights) {
		return errBackendWeightNotMatch
	}

	if len(a.Paths) != len(a.Methods) {
		return errPathMethodNotMatch
	}

	if a.LoadBalanceMethod == "" {
		a.LoadBalanceMethod = "rr"
		log.Printf("by default, app %s are using %s as load balance algorithm", a.Name, a.LoadBalanceMethod)
	}

	switch a.LoadBalanceMethod {
	case LBMWRR, LBMRR, LBMRandom:
		return nil
	default:
		return errBadLoadBalanceAlgorithm
	}
}

func getBalancer(loadBalanceMethod string, backends ...Backend) Balancer {
	switch loadBalanceMethod {
	case LBMWRR:
		return NewWRR(backends...)
	case LBMRR:
		return NewRR(backends...)
	case LBMRandom:
		return NewRdm(backends...)
	default:
		log.Panicf("bad load balance algorithm: %s", loadBalanceMethod)
		return nil // never here
	}
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	var config appConfig

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad configuration"))
		return
	}

	if err := checkAppConfig(&config); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad configuration: " + err.Error()))
		return
	}

	backends := []Backend{}
	for i, url := range config.Backends {
		backends = append(backends, NewBackend(url, config.Weights[i]))
	}

	balancer := getBalancer(config.LoadBalanceMethod, backends...)

	app := NewApp(balancer, !config.DisableTSR)

	for i, path := range config.Paths {
		app.AddRoute(path, strings.ToUpper(config.Methods[i]))
	}

	// replace breaker's map, FIXME: here may raise data race...
	breaker.apps[config.Name] = app

	w.Write([]byte("success!"))
}

func configManager() {
	http.HandleFunc("/app", appHandler)
	log.Fatal(http.ListenAndServe(":12345", nil))
}
