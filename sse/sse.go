package sse

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/hottguy/3go/log"
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

func Open(id string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string)
	put(id, ch)

	// 클라이언트 연결 종료 감지
	go func() {
		<-r.Context().Done()
		//ch가 끝났음을 마크
		log.Trace("클라이언트가 종료 %s, %+v", id, Clients)
	}()

	/*
		무한루프를 돌며 채널에 메세지가 도착하면 HTTP 응답을 함.
		즉, SSE 접속 1개당 1개의 고루틴이 계속 살아있는 것임.
		이 Open 함수가 리턴한다는 것은 해당 SSE 채널이 닫힌 것임.
	*/
	for msg := range ch {
		if err := r.Context().Err(); err != nil {
			// 브라우저 연결이 끊어짐
			fmt.Printf("Connection closed by client. %+v", err)
			remove(id)
			return
		}
		data := fmt.Sprintf("data: %s\n\n", msg)
		fmt.Fprint(w, data)
		w.(http.Flusher).Flush()
		log.Trace("메세지 전송됨. %v", id)
	}
}

/*
클라이언트 맵에 채널 추가
이미 추가된 채널이 있다면 닫기만 함.
왜냐하면 새 채널이 추가 되면 해당 채널은 댕글링 포인터가 되어 가비지컬렉션 대상이 됨.
*/
func put(id string, ch chan string) {
	cmu.Lock()
	defer cmu.Unlock()

	c, ok := Clients[id]
	if ok {
		close(c)
	}
	Clients[id] = ch
}

/*
클라이언트 맵의 특정 채널을 닫고 맵에서 삭제
*/
func remove(id string) {
	cmu.Lock()
	defer cmu.Unlock()

	c, ok := Clients[id]
	if ok {
		close(c)
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
