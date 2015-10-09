package garnet

import (
    "net"
    "log"
    "fmt"
    "io"
)

// Aggregator is a free-standing collector that reads from the Unix socket
type Aggregator struct {
    SocketPath string
    continueProcessing bool
}

func NewAggreagator(socket net.Listener) {
    return &Aggregator{
        Socket: socket,
        Done: false
    }
}

func (agggregator Aggregator) Start() net.Listener {
    for agggregator.Done {
        select {
        case <- agggregator.Done:
            log.Println("Ceasing to accept collector connections...")
            finishedChannel <- true
            return
        }

        log.Println("Accepting a collector connection...")
        data, err := readDataFromClient(socket)
        if err != nil {
            log.Printf("Error: %v\n", err)
        } else {
            log.Println(data)
        }
    }
}

func (agggregator *Aggregator) Stop(socketURL string) {
    agggregator.Done = true
    log.Println("Creating a mimic client to terminate socket accept thread")
    conn, err := net.Dial("unix", socketURL)
    if err != nil {
        log.Fatalf("Failed to open the final client connection to Garnet")
    }
    conn.Close()

}

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
