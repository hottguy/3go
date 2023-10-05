package sse

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	Clients = make(map[string]chan string)
	cmu     sync.Mutex
)

/*
해당 채널이 있다면 채널로 메세지 전송
*/
func Send(id, msg string) {
	c, ok := Clients[id]
	if ok {
		c <- msg
	}
}

/*
If this function return then sse connectoin will be closed.
*/
func Open(id string, w http.ResponseWriter, r *http.Request) {

	ch := make(chan string)
	put(id, ch)

	for msg := range ch {
		if err := r.Context().Err(); err != nil {
			remove(id)
			return
		}
		SendEvent(w, msg)
	}
}

func SendEvent(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	fmt.Fprintf(w, "data: %s\n\n", msg)
	w.(http.Flusher).Flush()
}

/*
클라이언트 맵에 채널 추가
이미 추가된 채널이 있다면 닫기만 함.
왜냐하면 새 채널이 추가 되면 해당 채널은 댕글링 포인터가 되어 가비지컬렉션 대상이 됨.
*/
func put(id string, ch chan string) {
	cmu.Lock()
	defer cmu.Unlock()

	oldch, ok := Clients[id]
	if ok {
		close(oldch)
	}
	Clients[id] = ch
}

/*
클라이언트 맵의 특정 채널을 닫고 맵에서 삭제
*/
func remove(id string) {
	cmu.Lock()
	defer cmu.Unlock()

	oldch, ok := Clients[id]
	if ok {
		close(oldch)
		delete(Clients, id)
	}
}

/*
모든 클라이언트로 향하는 채널을 닫음.
*/
func CloseAll() {
	for _, ch := range Clients {
		close(ch)
	}
}
