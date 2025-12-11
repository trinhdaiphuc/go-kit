package network

import (
	"context"
	"errors"
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
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	server.Shutdown(context.Background())
}
