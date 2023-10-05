package main

import (
	"net/http"
	"syscall"

	"github.com/hottguy/3go/app"
	"github.com/hottguy/3go/cfg"
	"github.com/hottguy/3go/log"
	"github.com/hottguy/3go/sse"
)

var (
	conf       = cfg.GetInstance("conf/config.json")
	fileServer = http.FileServer(http.Dir(conf.GetString("WebRoot")))
)

func main() {
	log.Initialize(
		conf.GetString("LogDir"),
		conf.GetString("LogFileNamePattern"),
		conf.GetString("LogLevel"),
	)

	log.Trace("config.json loaded. %+v", conf)

	app.RegSignalCallback(syscall.SIGTERM, close)
	app.RegSignalCallback(syscall.SIGINT, close)
	app.RegSignalCallback(syscall.SIGUSR1, log.Rotate)
	app.Run([]*app.App{
		app.Http(app.Svr(conf.GetString("Http"), mux)),
	})
}

func close() {
	sse.CloseAll()
	log.Trace("모든 채널을 닫음. %+v", sse.Clients)
}

func mux(w http.ResponseWriter, r *http.Request) {
	log.Trace("%+v", r.URL)

	switch r.URL.Path {
	case "/events":
		id := r.URL.Query().Get("id")
		sse.Open(id, w, r) //이 함수는 채널이 닫힐 때 까지 반환되지 않음.
	case "/send":
		id := r.URL.Query().Get("id")
		sse.Send(id, `{"name":"강아지", "age":20}`)
	default:
		fileServer.ServeHTTP(w, r)
	}
}
