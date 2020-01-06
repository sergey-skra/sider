package main

import (
	"testing"

	"github.com/sergebraun/sider/pkg/pb"
	"golang.org/x/net/context"
)

func TestSet(t *testing.T) {
	testServer := initServer()
	req := &pb.SetRequest{
		Key:     "foo",
		Value:   []byte("value"),
		Expires: -1,
	}
	_, err := testServer.Set(context.Background(), req)

	if err != nil {
		t.Error(err)
	}
}

func initServer() *Server {
	return &Server{
		db: Open(),
	}
}
