package garnet

import (
    "net"
    "log"
    "os"
    "os/signal"
    "io/ioutil"
    "syscall"
    "path/filepath"
    "time"
    "encoding/json"
)

const socketPath = "/tmp/garnet.sock"
const configDir = "config"

// Config holds the JSON configuration of each config file under the configDir
type Config struct {
        Command string `json:"command"`
        Interval int `json:"interval"`
}

// Signal handler catches SIGINT and SIGTERM and sends a "done" flag to the main loop
func signalHandler(signalChannel chan os.Signal, signalDoneChannel chan bool){
    // block the goroutine until we get a signal
    signal := <-signalChannel
    log.Printf("Got signal %v, exiting...\n", signal)
    // Send the message to terminate the app
    signalDoneChannel <- true
}

// Connect to the socket as a client to unblock the Accept call
func mimicFinalClient(socketURL string) {
    log.Println("Creating a mimic client to terminate socket accept thread")
    conn, err := net.Dial("unix", socketURL)
    if err != nil {
        log.Fatalf("Failed to open the final client connection to Garnet")
    }
    conn.Close()
}

// Create a list of collectors from all the config files in configDir
func collectorsFromConfig(configDir string) []*Collector {
    globs, err := filepath.Glob(filepath.Join(configDir, "*.json"))
    if err != nil {
        log.Fatalf("Failed to search for configs in: %s, reason: %v", configDir, err)
    }

    totalConsumers := len(globs)
    collectors := make([]*Collector, totalConsumers)

    for idx, path := range globs {
        file, err := os.Open(path)
        if err != nil {
            log.Fatalf("Failed to open file: %s, reason: %v", path, err)
        }

        bytes, err := ioutil.ReadAll(file)
        if err != nil {
            log.Fatalf("Failed to read file: %s, reason: %v", path, err)
        }

        var cfg Config
        err = json.Unmarshal(bytes, &cfg)
        if err != nil {
            log.Fatalf("Failed to parse JSON file: %s, reason: %v", path, err)
        }

        collector := NewCollector(path, cfg.Command, time.Duration(cfg.Interval) * time.Second)
        collectors[idx] = collector
    }
    return collectors
}

// Run is the main entry point into garnet, packaged up behind a nice little func
func Run() {
    // Create a channel to pass to os.Notify for OS signal handling
    signalChannel := make(chan os.Signal, 1)
    signalDoneChannel := make(chan bool, 1)
    signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
    go signalHandler(signalChannel, signalDoneChannel)

    socket, err := net.Listen("unix", socketPath)
    if err != nil {
        log.Fatalf("Failed to create a new Unix socket: err: %v\n", err)
    }
    defer socket.Close()

    log.Printf("Opened a socket connection '%s'\n", socketPath)

    // Start the aggregation collector
    aggregationDoneChannel := make(chan bool, 1)
    aggregationCleanUpChannel := make(chan bool, 1)
    go AggregationWorker(socket, aggregationDoneChannel, aggregationCleanUpChannel)

    // Fire a "tick" through the channel every 5 seconds
    collectors := collectorsFromConfig(configDir)
    for _, collector := range collectors {
        go collector.Start(socketPath)
    }

    // Wait until we get a catchable signal before cleaning up
    <- signalDoneChannel

    // Stop sending new ticks to the collector launching goroutine
    for _, collector := range collectors {
        log.Printf("Stopping collector '%s' ...", collector.Name)
        collector.Stop()
    }

    // Tell the collector aggregator to stop processing connections
    aggregationDoneChannel <- true
    mimicFinalClient(socketPath)
    <- aggregationCleanUpChannel
}
