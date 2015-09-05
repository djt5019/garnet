package main

import (
    "net"
    "log"
    "os"
    "os/signal"
    "syscall"
)


func signalHandler(signalChannel chan os.Signal, doneChannel chan bool){
    // block the goroutine until we get a signal
    signal := <-signalChannel
    log.Printf("Got signal %v, exiting...\n", signal)
    // Send the message to terminate the app
    doneChannel <- true
}

func main() {
    // Create a channel to pass to os.Notify for OS signal handling
    signalChannel := make(chan os.Signal, 1)
    signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

    doneChannel := make(chan bool, 1)
    go signalHandler(signalChannel, doneChannel)

    socket, err := net.Listen("unix", "/tmp/garnet.sock")
    if err != nil {
        log.Fatalf("Failed to create a new Unix socket: err: %v\n", err)
    }
    defer socket.Close()

    log.Printf("Opened a socket connection '/tmp/garnet.sock'\n")

    <- doneChannel
}
