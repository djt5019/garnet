package main

import (
    "os"
    "net"
    "log"
)

func main() {
    if len(os.Args) != 2 {
        log.Fatalf("Missing required Unix socket URL argument")
    }

    log.Printf("Connecting to %s", os.Args[1])
    connection, err := net.Dial("unix", os.Args[1])
    if err != nil {
        log.Fatalf("Failed to open the Unix socket: err: %v", err)
    }

    defer connection.Close()

    _, err = connection.Write([]byte("application.prog.counters|1c"))
    if err != nil {
        log.Fatalf("Failed to write to the Unix socket: err: %v", err)
    }
}
