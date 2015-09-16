package garnet

import (
    "net"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
)

const socketPath = "/tmp/garnet.sock"


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
    collectors := make([]*Collector, 5)
    for i := 0; i < 5; i++ {
        collector := NewCollector("./sampleCollector", 5 * time.Second)
        collectors[i] = collector
        go collector.Start(socketPath)
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
