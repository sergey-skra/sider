package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sergebraun/sider/pkg/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	grpcAddress = flag.String("grpcAddress", "localhost:7777", "The server address in the format of host:port")
	restAddress = flag.String("restAddress", "localhost:7778", "The server address in the format of host:port")
	certFile    = "cert/server.crt"
	keyFile     = "cert/server.key"
)

// Server represents the gRPC server.
type Server struct {
	db *DB
}

// Set key to hold the value.
func (s *Server) Set(c context.Context, r *pb.SetRequest) (*pb.SetResponse, error) {
	s.db.Set(r.Key, Entry{r.Value, r.Expires})
	return &pb.SetResponse{}, nil
}

// Get the entry from DB.
func (s *Server) Get(c context.Context, r *pb.GetRequest) (*pb.GetResponse, error) {
	e, ok := s.db.Get(r.Key)
	if !ok {
		return &pb.GetResponse{}, errors.New("key doesn't exist")
	}
	res := &pb.GetResponse{
		Value:   e.Value,
		Expires: e.Expires,
	}
	return res, nil
}

// Update the entry in DB.
func (s *Server) Update(c context.Context, r *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	err := s.db.Update(r.Key, r.Value)
	if err != nil {
		return &pb.UpdateResponse{}, errors.New("key doesn't exist")
	}
	return &pb.UpdateResponse{}, nil
}

// Remove delete key.
func (s *Server) Remove(c context.Context, r *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	s.db.Remove(r.Key)
	return &pb.RemoveResponse{}, nil
}

// Keys returns all keys mathing the pattern.
func (s *Server) Keys(c context.Context, r *pb.KeysRequest) (*pb.KeysResponse, error) {
	arr := s.db.Keys(r.Pattern)
	return &pb.KeysResponse{Key: arr}, nil
}

// TTL returns time-to-live value for the key.
func (s *Server) TTL(c context.Context, r *pb.TTLRequest) (*pb.TTLResponse, error) {
	t := s.db.TTL(r.Key)
	return &pb.TTLResponse{Time: t}, nil
}

// private type for Context keys
type contextKey int

const clientIDKey contextKey = 0

func credMatcher(headerName string) (mdName string, ok bool) {
	if headerName == "Login" || headerName == "Password" {
		return headerName, true
	}
	return "", false
}

// authenticateAgent check the client credentials
func authenticateClient(ctx context.Context, s *Server) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		clientLogin := strings.Join(md["login"], "")
		clientPassword := strings.Join(md["password"], "")
		if clientLogin != "john" {
			return "", fmt.Errorf("unknown user %s", clientLogin)
		}
		if clientPassword != "doe" {
			return "", fmt.Errorf("bad password %s", clientPassword)
		}
		log.Printf("authenticated client: %s", clientLogin)
		return "42", nil
	}
	return "", fmt.Errorf("missing credentials")
}

// unaryInterceptor calls authenticateClient with current context
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	s, ok := info.Server.(*Server)
	if !ok {
		return nil, fmt.Errorf("unable to cast server")
	}
	clientID, err := authenticateClient(ctx, s)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, clientIDKey, clientID)
	return handler(ctx, req)
}

func startGRPCServer(address, certFile, keyFile string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	s := Server{Open()}

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("could not load TLS keys: %s", err)
	}

	// Create an array of gRPC options with the credentials
	opts := []grpc.ServerOption{grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryInterceptor)}

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterSiderServer(grpcServer, &s)

	fmt.Println("starting server...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %s", err)
	}
	return nil
}

func startRESTServer(address, grpcAddress, certFile string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(credMatcher))

	creds, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		return fmt.Errorf("could not load TLS certificate: %s", err)
	}
	// Setup the client gRPC options
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	// Register
	err = pb.RegisterSiderHandlerFromEndpoint(ctx, mux, grpcAddress, opts)
	if err != nil {
		return fmt.Errorf("could not register service Ping: %s", err)
	}
	log.Printf("starting HTTP/1.1 REST server on %s", address)
	http.ListenAndServe(address, mux)
	return nil
}

// main start a gRPC server and waits for connection
func main() {
	flag.Parse()
	// fire the gRPC server in a goroutine
	go func() {
		err := startGRPCServer(*grpcAddress, certFile, keyFile)
		if err != nil {
			log.Fatalf("failed to start gRPC server: %s", err)
		}
	}()
	// fire the REST server in a goroutine
	go func() {
		err := startRESTServer(*restAddress, *grpcAddress, certFile)
		if err != nil {
			log.Fatalf("failed to start REST server: %s", err)
		}
	}()
	select {}
}
