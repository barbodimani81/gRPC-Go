package grpcutils

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	serviceName     string
	ServerGRPC      *grpc.Server
	RegisterService func(ctx context.Context, grpcSrv *Server, closer *Closer) error
}

type Closer struct {
	closers []func() error
}

func (c *Closer) Add(fn func() error) {
	c.closers = append(c.closers, fn)
}

func (c *Closer) Close() {
	for _, fn := range c.closers {
		if err := fn(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}
}

func NewServer(serviceName string, disableHealthCheck bool) *Server {
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(30*1024*1024), // 30MB
		grpc.MaxSendMsgSize(30*1024*1024),
	)

	return &Server{
		serviceName:     serviceName,
		ServerGRPC:      grpcServer,
		RegisterService: nil,
	}
}

func (s *Server) UseUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) {
	s.ServerGRPC = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.MaxRecvMsgSize(30*1024*1024),
		grpc.MaxSendMsgSize(30*1024*1024),
	)
}

func (s *Server) Run(ctx context.Context, address string) {
	closer := &Closer{}
	defer closer.Close()

	// Register service
	if s.RegisterService != nil {
		if err := s.RegisterService(ctx, s, closer); err != nil {
			log.Fatalf("Failed to register service: %v", err)
		}
	}

	// Enable reflection if env var is set
	if os.Getenv("GRPC_REFLECTION_ENABLED") == "true" {
		reflection.Register(s.ServerGRPC)
		log.Println("gRPC reflection enabled")
	}

	// Start listening
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("%s gRPC server starting on %s", s.serviceName, address)

	// Start server in goroutine
	go func() {
		if err := s.ServerGRPC.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down gracefully...")
	s.ServerGRPC.GracefulStop()
	log.Println("Server stopped")
}
