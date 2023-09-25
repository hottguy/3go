package main

import (
	"3go/app"
	"3go/log"
	"net/http"
	"syscall"
)

func main() {
	app.RegSignalCallback(syscall.SIGUSR1, log.Rotate)
	app.Run([]*app.App{
		app.Http(app.Svr(":5000", mux)),
	})
}

func mux(w http.ResponseWriter, r *http.Request) {
	log.Trace("%v", "asdfasdf")
}
