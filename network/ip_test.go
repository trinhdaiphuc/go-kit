package network

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
)

func TestLocalIP(t *testing.T) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf("listen: %v\n", err)
	}

	server := http.Server{}
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Fatalf("serve: %v\n", err)
		}
	}()

	defer server.Shutdown(context.Background())
	defer listener.Close()
	ip, err := LocalIP()
	fmt.Println(ip)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("local ip: %s", ip)
}
