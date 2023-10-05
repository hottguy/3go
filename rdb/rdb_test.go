package rdb

import (
	"log"
	"testing"
)

func TestXxx(t *testing.T) {
	rdbx := GetInstance()
	rdbx.Open("mysql", "activar:dorxlqj#2023@tcp(49.247.4.87:33333)/heyid")
	defer rdbx.Close()

	r, err := rdbx.Call("id_select", "1")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	log.Printf("%+v", r)
}
