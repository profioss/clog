package main

import (
	"fmt"
	"log"
	"os"
	"github.com/profioss/clog"
)

func main() {
	fname := fmt.Sprintf("/tmp/clog-example-%d.log", os.Getpid())

	fd, err := clog.OpenFile(fname) // use provided function for proper file opening (append mode etc.)
	if err != nil {
		log.Fatal("Log file error:", err)
	}
	defer fd.Close() // don't forget to flush & close storage facility

	// available log levels: disabled | error | warning | info | debug
	logger, err := clog.New(fd, "info", false)
	// logger, err := clog.New(fd, "debug", false)
	// logger, err := clog.New(fd, "disabled", false) // DisabledLevel is good for testing, benchmarking etc.
	if err != nil {
		log.Fatal("Logger error:", err)
	}

	logger.Info("this message is shown because we are in clog.InfoLevel")
	logger.Debug("this message is not shown because we are not in clog.DebugLevel")
	logger.Error("this message is stored in log and printed to stderr")
	logger.Fatalf("this message is fatal, program exits immediately with error code 1.\ncheck log file: %s\nbye!", fname)

	fmt.Println("not shown")
}
