package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sergebraun/sider/pkg/pb"
	"github.com/sergebraun/sider/pkg/util"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	ishell "gopkg.in/abiosoft/ishell.v2"
)

var serverAddr = flag.String("server_addr", "localhost:7777", "The server address in the format of host:port")

// TODO:
// // Client represents a client connected to the Sider server.
// type Client struct {
// 	// Username is a user name for authentication.
// 	Username string
// 	// Password is a password for authentication.
// 	Password string
// }

// Auth holds the login/password
type Auth struct {
	Login    string
	Password string
}

// GetRequestMetadata gets the current request metadata
func (a *Auth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"login":    a.Login,
		"password": a.Password,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (a *Auth) RequireTransportSecurity() bool {
	return true
}

func main() {
	flag.Parse()

	// Create the client TLS credentials
	creds, err := credentials.NewClientTLSFromFile("cert/server.crt", "")
	if err != nil {
		log.Fatalf("could not load tls cert: %s", err)
	}

	// Setup the login/pass
	auth := Auth{
		Login:    "john",
		Password: "doe",
	}

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewSiderClient(conn)

	shell := ishell.New()
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "Set key to hold the value.",
		Func: func(c *ishell.Context) {
			if len(c.Args) <= 1 {
				fmt.Println("ERR wrong number of arguments for 'set' command")
				return
			}
			client.Set(context.TODO(), &pb.SetRequest{
				Key:     c.Args[0],
				Value:   util.Itob(c.Args[1]),
				Expires: -1,
			})
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "Get the value of key.",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				fmt.Println("ERR wrong number of arguments for 'get' command")
				return
			}
			val, err := client.Get(context.TODO(), &pb.GetRequest{Key: c.Args[0]})
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(util.Btoi(val.Value))
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "keys",
		Help: "Returns all keys matching pattern.",
		Func: func(c *ishell.Context) {
			// client.Keys(context.TODO(), ...)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "remove",
		Help: "Removes the specified key.",
		Func: func(c *ishell.Context) {
			// client.Remove(context.TODO(), ...)
		},
	})

	shell.Run()
}
