package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"runtime"
	"syscall"
)

var signalCallback = map[os.Signal]func(){}

func RegSignalCallback(sig os.Signal, cb func()) {
	signalCallback[sig] = cb
}

func Svr(addr string, h http.HandlerFunc) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: h,
	}
}

// App 구조체
type App struct {
	srv  *http.Server
	cert string
	key  string
}

// Http App 생성
func Http(srv *http.Server) *App {
	return &App{
		srv: srv,
	}
}

// Https App 생성
// addr 은 할당IP:Port 형태의 주소 :Port는 모든 IP에 대한 Listen
// 멀티 인증서(SAN) 혹은 와일드카드 인증서(*.domain) 로 멀티 도메인 지원
func Https(
	srv *http.Server,
	cert, key string,
) *App {
	srv.TLSConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		//MaxVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
			tls.CurveP521,
		},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	return &App{
		srv:  srv,
		cert: cert,
		key:  key,
	}
}

// []*app.AppX 를 인자로 복수개의 서버를 실행할 수 있음.
// 서버목록을 차례로 실행하고
// 정상 종료를 위한 시그널 대기
func Run(apps []*App) {

	for _, a := range apps {
		//주소없는 서버는 실행할 수 없음.
		if a.srv.Addr == "" {
			continue
		}
		/*
			수행순서는 아래 한 줄 주석에 상세히 설명함.
		*/
		go a.Go()      // 1. LinstenAndServe() 때문에 이 고루틴은 블록상태
		defer a.Stop() // 3. defer 에 걸어놓은 Stop()은 Wait() 함수종료 후 실행됨.
	}
	Wait() // 2. Go() 함수를 고루틴으로 띄우고 이 함수에서 시그널 대기
}

// 서버시작
func (a *App) Go() {
	var err error
	log.Print("App Started. " + a.String())
	if a.cert == "" || a.key == "" {
		err = a.srv.ListenAndServe()
	} else {
		err = a.srv.ListenAndServeTLS(a.cert, a.key)
	}
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}

// 서버종료
func (a *App) Stop() {
	err := a.srv.Shutdown(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	log.Print("App Stopped. " + a.String())
}

/*
App 구조체의 문자열 표현
*/
func (a *App) String() string {
	fname := GetFuncName(a.srv.Handler)
	return fmt.Sprintf("%s %s %s %s", a.srv.Addr, fname, a.cert, a.key)
}

// 함수 포인터로 부터 함수명 반환
func GetFuncName(pointer any) string {
	p := reflect.ValueOf(pointer).Pointer()
	f := runtime.FuncForPC(p).Name()
	return filepath.Base(f)
}

/*
정상종료를 위한 시그널 대기

SIGINT(2) 혹은 SIGTERM(15) 시그널을 수신하면 정상종료하고 나머지 시그널은 무시한다.
수신한 시그널은 INFO 레벨 로그를 남긴다. 다만, SIGURG(23), SIGWINCH(28) 는 너무 많이
발생하므로 로그를 남기지 않는다.

	SIGHUP		1	Hangup (POSIX)	Terminate
	SIGINT		2	Terminal interrupt (ANSI)	Terminate
	SIGQUIT		3	Terminal quit (POSIX)	Core Dump
	SIGILL		4	Illegal instruction (ANSI)	Core Dump
	SIGTRAP		5	ctx.Trace trap (POSIX)	Core Dump
	SIGABRT		6	Aborted	Core Dump
	SIGBUS		7	BUS error (4.2 BSD)	Core Dump
	SIGFPE		8	Floating point exception (ANSI)
	SIGKILL		9	Kill(can't be caught or ignored) (POSIX)	Terminate
	SIGUSR1		10	User defined signal 1 (POSIX)
	SIGSEGV		11	Invalid memory segment access (ANSI)	Terminate + Core Dump
	SIGUSR2		12	User defined signal 2 (POSIX)
	SIGPIPE		13	Write on a pipe with no reader, Broken pipe (POSIX)
	SIGALRM		14	Alarm clock (POSIX)
	SIGTERM		15	Termination (ANSI)	Terminate
	SIGSTKFLT	16	Stack fault
	SIGCHLD		17	Child process has stopped or exited, changed (POSIX)	Ignore
	SIGCONTv	18	Continue executing, if stopped (POSIX)	Restart
	SIGSTOP		19	Stop executing(can't be caught or ignored) (POSIX)	Suspend
	SIGTSTP		20	Terminal stop signal (POSIX)	Suspend
	SIGTTIN		21	Background process trying to read, from TTY (POSIX)
	SIGTTOU		22	Background process trying to write, to TTY (POSIX)
	SIGURG		23	Urgent condition on socket (4.2 BSD)
	SIGXCPU		24	CPU limit exceeded (4.2 BSD)
	SIGXFSZ		25	File size limit exceeded (4.2 BSD)
	SIGVTALRM	26	Virtual alarm clock (4.2 BSD)
	SIGPROF		27	Profiling alarm clock (4.2 BSD)
	SIGWINCH	28	Window size change (4.3 BSD, Sun)
	SIGIO		29	I/O now possible (4.2 BSD)	Terminate
	SIGPWR		30	Power failure restart (System V)
*/
func Wait() {

	ch := make(chan os.Signal, 1)
	signal.Notify(ch)

	for {
		sig := <-ch

		/*
			수신한 모든 시그널을 기록으로 남김.
			SIGURG(23), SIGWINCH(28) 는 너무 많이 발생하므로 로그 남기지 않음
			SIGCHLD(17) butler.Info() 함수가 grep 명령을 사용할 때 발생 (무시)
		*/
		sig17 := sig == syscall.Signal(17)
		sig23 := sig == syscall.Signal(23)
		sig28 := sig == syscall.Signal(28)

		f, ok := signalCallback[sig]
		if ok {
			f()
		}

		if !sig17 && !sig23 && !sig28 && !ok {
			log.Printf("Signal: %s (%d)", sig.String(), sig)
		}
		/*
			인터럽트시그널 혹은 종료시그널인 경우 종료
			SIGINT	2	Terminal interrupt (ANSI)
			SIGTERM	15	Termination (ANSI)
		*/
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			break
		}
	}
}
