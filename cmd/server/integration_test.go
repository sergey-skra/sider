// +build integration

package main

import (
	"flag"
	"log"
	"testing"

	"github.com/sergebraun/sider/pkg/pb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var serverAddr = flag.String("server_addr", "localhost:7777", "The server address in the format of host:port")

// Authentication holds the login/password
type Authentication struct {
	Login    string
	Password string
}

// GetRequestMetadata gets the current request metadata
func (a *Authentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"login":    a.Login,
		"password": a.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (a *Authentication) RequireTransportSecurity() bool {
	return true
}

func TestServer(t *testing.T) {
	flag.Parse()

	// Create the client TLS credentials
	creds, err := credentials.NewClientTLSFromFile("cert/server.crt", "")
	if err != nil {
		log.Fatalf("could not load tls cert: %s", err)
	}

	// Setup the login/pass
	auth := Authentication{
		Login:    "john",
		Password: "doe",
	}

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewSiderClient(conn)

	client.Set(context.TODO(), &pb.SetRequest{
		Key:     "foo",
		Value:   []byte("val"),
		Expires: -1,
	})
	client.Get(context.TODO(), &pb.GetRequest{Key: "foo"})
}
