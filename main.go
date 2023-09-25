package main

import (
	"3go/customlog"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	customlog.SetLogLevel(customlog.INFO)
	customlog.SetOutput(file)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP)
	<-sigCh
	fmt.Println("Received SIGUSR1. 출력: ㅌㅌㅌ")

	for {
		customlog.Trace("This is a trace message")
		customlog.Debug("This is a debug message")
		customlog.Info("This is an info message")
		customlog.Warning("This is a warning message")
		customlog.Error("This is an error message")
		customlog.Fatal("This is an 치명적인 error message")
		time.Sleep(1000 * time.Millisecond)
	}
}
