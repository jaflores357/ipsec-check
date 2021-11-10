package main

import (
    "fmt"
    "io/ioutil"
    "strings"
    "net/http"
    "os"
    "log"
    "gopkg.in/yaml.v2"
)

type Config struct {
    Slave       string `yaml:"slave"`
    Port        string `yaml:"port"`
    StatusFile  string `yaml:"statusFile"`
    LogFile     string `yaml:"logFile"`
}

func api(w http.ResponseWriter, req *http.Request) {
    slaveStatus := checkSlave(config.Slave)    
    status := func() int { if slaveStatus == 200  { return http.StatusServiceUnavailable } else { return http.StatusOK } }()
    dat, err := ioutil.ReadFile(config.StatusFile)
    if err == nil {
        if strings.Contains(string(dat), "UP") {
            status = http.StatusOK
        }
    }
    log.Printf("Local statusCode %d, Remote statusCode %d", status, slaveStatus)
    w.WriteHeader(status)    
}

func heartbeat(w http.ResponseWriter, req *http.Request) {
    status := http.StatusServiceUnavailable
    dat, err := ioutil.ReadFile(config.StatusFile)
    if err == nil {
        if strings.Contains(string(dat), "UP") {
            status = http.StatusOK
        }
    }
    log.Printf("Heartbeat statusCode %d", status)
    w.WriteHeader(status)    
}



func checkSlave (req string) int {
    client := &http.Client{}
    request, err := http.NewRequest("GET", req, nil)
    response, err := client.Do(request)
    if err != nil {
        log.Print("Remote error: ", err)
        return http.StatusServiceUnavailable
    }
    defer response.Body.Close()
    return response.StatusCode
}

func processError(err error) {
    fmt.Println(err)
    os.Exit(2)
}

func readConfig(cfg *Config) {
    f, err := os.Open("/etc/ipsec-check.yaml")
    if err != nil {
        processError(err)
    }
    defer f.Close()

    decoder := yaml.NewDecoder(f)
    err = decoder.Decode(cfg)
    if err != nil {
        processError(err)
    }
} 

var config Config

func main() {

    file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }

    defer file.Close()

    log.SetOutput(file)
    log.Print("Starting application!")

    readConfig(&config)
    http.HandleFunc("/api", api)
    http.HandleFunc("/heartbeat", heartbeat)
    log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}


