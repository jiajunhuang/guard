package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	errNameEmpty               = errors.New("name is required")
	errBackendWeightNotMatch   = errors.New("backend and weight does not match")
	errPathMethodNotMatch      = errors.New("path and method does not match")
	errBadLoadBalanceAlgorithm = errors.New("bad load balance algorithm, only wrr, rr, random are support now")

	configSync = make(chan appConfig)
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

func getAPP(config *appConfig) *Application {
	backends := []Backend{}
	for i, url := range config.Backends {
		backends = append(backends, NewBackend(url, config.Weights[i]))
	}

	balancer := getBalancer(config.LoadBalanceMethod, backends...)

	app := NewApp(balancer, !config.DisableTSR)

	for i, path := range config.Paths {
		app.AddRoute(path, strings.ToUpper(config.Methods[i]))
	}

	return app
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

	// replace breaker's map, FIXME: here may raise data race...
	breaker.apps[config.Name] = getAPP(&config)

	go func() { configSync <- config }()
	w.Write([]byte("success!"))
}

type breakerConfig struct {
	APPs map[string]appConfig `json:"apps"`
}

func readFromFile(path string) ([]byte, error) {
	f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

func configKeeper() {
	// first try to load config
	b := breakerConfig{make(map[string](appConfig))}

	fileBytes, err := readFromFile(*configPath)
	if err != nil {
		log.Panicf("failed to use config file %s: %s", *configPath, err)
	}
	if err := json.Unmarshal(fileBytes, &b); err == nil {
		log.Printf("loading config from config file")
		for k, v := range b.APPs {
			breaker.apps[k] = getAPP(&v)
		}
	} else {
		log.Printf("failed to unmarshal config file %s because %s", *configPath, err)
	}

	// listen channel for sync
	for config := range configSync {
		f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			log.Panicf("failed to open config file: %s", err)
		}

		if err := checkAppConfig(&config); err != nil {
			log.Printf("receive a bad config: %+v, ignore it", config)
			continue
		}

		if err = json.Unmarshal(fileBytes, &b); err != nil && len(fileBytes) > 0 {
			log.Printf("failed to unmarshal config file %s because %s", *configPath, err)
			continue
		}
		b.APPs[config.Name] = config

		f.Truncate(0)
		f.Seek(0, 0)
		jsonBytes, err := json.Marshal(b)
		if err != nil {
			log.Printf("failed to marshal configuration, err is: %s", err)
			continue
		}
		_, err = f.Write(jsonBytes)
		if err != nil {
			log.Printf("failed to sync configuration to backup file %s because: %s", *configPath, err)
			continue
		}

		f.Close()
		log.Printf("sync configuration to backup file %s succeed", *configPath)
	}

	log.Printf("stop sync config file")
}

func configIndexHandler(w http.ResponseWriter, r *http.Request) {
	if fileBytes, err := readFromFile(*configPath); err != nil {
		log.Printf("failed to read from %s: %s", *configPath, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(fileBytes)
	}
}

func configManager() {
	go configKeeper()
	http.HandleFunc("/app", appHandler)
	http.HandleFunc("/", configIndexHandler)
	log.Fatal(http.ListenAndServe(*configAddr, nil))
}
