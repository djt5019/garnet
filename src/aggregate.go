package garnet

import (
    "net"
    "log"
    "fmt"
    "io"
)

// AggregationWorker collects data from the socket and sends it to where it needs to go
func AggregationWorker(socket net.Listener, doneChannel, finishedChannel chan bool) {
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
