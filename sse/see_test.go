package sse

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSSE(t *testing.T) {
	// 가상의 HTTP 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SendEvent(w, "안녕?")
	}))
	defer server.Close()

	// 서버 주소를 이용한 HTTP 요청 보내기
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("HTTP GET 요청 실패: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("HTTP 응답 body 읽기 실패: %v", err)
	}
	t.Log(string(body))
}
