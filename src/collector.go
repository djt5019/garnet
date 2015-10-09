package garnet

import (
    "time"
    "os/exec"
    "log"
)

// Collector struct which contains the command to execute and a timer
type Collector struct {
    Name string
    Command string
    Ticker *time.Ticker
}

// NewCollector creates a new instance of a collector
func NewCollector(name, command string, d time.Duration) *Collector {
    return &Collector{
        Name: name,
        Command: command,
        Ticker: time.NewTicker(d),
    }
}

// Stop halts the collectors timer and allows it to exit
func (collector Collector) Stop() {
    collector.Ticker.Stop()
}

// Start begins execution of the collectors command every "duration" seconds.
func (collector Collector) Start(SocketURL string) {
    // This loop executes every time the ticker sends a "tick" message
    // to the channel after "duration" seconds.
    for range collector.Ticker.C {
        cmd := exec.Command(collector.Command, SocketURL)
        err := cmd.Run()
        if err != nil {
            log.Printf("Failed to invoke collector %s, reason: %v", collector.Command, err)
        }
    }
}
