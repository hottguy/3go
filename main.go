package main

import (
	"net/http"
	"syscall"

	"github.com/hottguy/3go/app"
	"github.com/hottguy/3go/log"
)

func main() {
	app.RegSignalCallback(syscall.SIGUSR1, log.Rotate)
	app.Run([]*app.App{
		app.Http(app.Svr(":5000", mux)),
	})
}

func mux(w http.ResponseWriter, r *http.Request) {
	log.Trace("%+v", r.URL)
}
