package main

import (
    "fmt"
    "net"
    "log"
    "os"
    "os/signal"
    "os/exec"
    "syscall"
    "io"
    "time"
)

const socketPath = "/tmp/garnet.sock"

type Collector struct {
    Command string
    Ticker *time.Ticker
}

func NewCollector(command string, d time.Duration) Collector {
    return Collector{
        Command: command,
        Ticker: time.NewTicker(d),
    }
}

func (collector *Collector) Stop() {
    collector.Ticker.Stop()
}

// Signal handler catches SIGINT and SIGTERM and sends a "done" flag to the main loop
func signalHandler(signalChannel chan os.Signal, doneChannel chan bool){
    // block the goroutine until we get a signal
    signal := <-signalChannel
    log.Printf("Got signal %v, exiting...\n", signal)
    // Send the message to terminate the app
    doneChannel <- true
}

// Collects data from the socket and sends it to where it needs to go
func aggregateCollectorData(socket net.Listener, doneChannel, finishedChannel chan bool) {
    for {
        select {
        case <- doneChannel:
            log.Println("Ceasing to accept collector connections...")
            finishedChannel <- true
            return
        default:
            log.Println("Accepting a collector connection...")
            data, err := readDataFromClient(socket)
            if err != nil {
                log.Printf("Error: %v\n", err)
            } else {
                log.Println(data)
            }
        }
    }
}

// Accept a connection from the socket and read 512 bytes of data into a buffer
func readDataFromClient(socket net.Listener) ([]byte, error) {
    readBuffer := make([]byte, 512)

    fd, err := socket.Accept()
    if err != nil {
        return nil, fmt.Errorf("Failed to accept a connection: err: %v\n", err)
    }

    defer fd.Close()

    bytesRead, err := fd.Read(readBuffer)

    if err == io.EOF {
        return nil, fmt.Errorf("Received no data from client connection")
    }

    if err != nil {
        return nil, fmt.Errorf("Failed to read from the socket into the buffer: err: %v\n", err)
    }

    return readBuffer[:bytesRead], nil
}

// Connect to the socket as a client to unblock the Accept call
func mimicFinalClient(socketUrl string) {
    log.Println("Creating a mimic client to terminate socket accept thread")
    conn, err := net.Dial("unix", socketUrl)
    if err != nil {
        log.Fatalf("Failed to open the final client connection to Garnet")
    }
    conn.Close()
}

func launchCollector(collector *Collector, socketUrl string) {
    // This loop executes every time the ticker sends a "tick" message
    // to the channel after "duration" seconds.
    for range collector.Ticker.C {
        cmd := exec.Command(collector.Command, socketUrl)
        err := cmd.Run()
        if err != nil {
            log.Printf("Failed to invoke collector %s, reason: %v", collector.Command, err)
        }
    }
}

func main() {
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
    go aggregateCollectorData(socket, aggregationDoneChannel, aggregationCleanUpChannel)

    // Fire a "tick" through the channel every 5 seconds
    collectors := make([]Collector, 5)
    for i := 0; i < 5; i++ {
        collector := NewCollector("./sampleCollector", 5 * time.Second)
        collectors[i] = collector
        go launchCollector(&collector, socketPath)
    }

    // Wait until we get a catchable signal before cleaning up
    <- signalDoneChannel

    // Stop sending new ticks to the collector launching goroutine
    for _, collector := range collectors {
        log.Printf("Stopping...")
        collector.Stop()
    }

    // Tell the collector aggregator to stop processing connections
    aggregationDoneChannel <- true
    mimicFinalClient(socketPath)
    <- aggregationCleanUpChannel
}
