package main

import (
    "fmt"
    "io/ioutil"
    "strings"
    "net/http"
    "os"
    "log"
    "time"
    "gopkg.in/yaml.v2"
    "github.com/lestrrat-go/file-rotatelogs"
)

type Config struct {
    Slave           string `yaml:"slave"`
    Port            string `yaml:"port"`
    StatusFile      string `yaml:"statusFile"`
    LogPrefix       string `yaml:"logPrefix"`
    logRotationTime int    `yaml:"logRotationTime"`
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

func initiLogger() {
    path := config.LogPrefix
    writer, err := rotatelogs.New(
        fmt.Sprintf("%s.%s", path, "%Y%m%d%H%M"),
        rotatelogs.WithLinkName(config.LogPrefix + ".log"),
        rotatelogs.WithRotationCount(2),
        rotatelogs.WithRotationTime(time.Second*time.Duration(config.logRotationTime)),
    )
    if err != nil {
        log.Fatalf("Failed to Initialize Log File %s", err)
    }
    log.SetOutput(writer)
    return
}

var config Config

func main() {

    readConfig(&config)
    initiLogger()

    log.Print("Starting application!")
    
    http.HandleFunc("/api", api)
    http.HandleFunc("/heartbeat", heartbeat)
    log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}


