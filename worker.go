package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	pb "grpc-project/proto"

	"google.golang.org/grpc"
)

type Worker struct {
	pb.UnimplementedWorkerServer

	registry pb.RegistryClient
}

func (w *Worker) Echo(ctx context.Context, t *pb.Text) (*pb.Text, error) {
	return &pb.Text{Msg: t.Msg}, nil
}

func (w *Worker) Add(ctx context.Context, n *pb.Numbers) (*pb.Result, error) {
	return &pb.Result{Value: n.A + n.B}, nil
}

// Usa il contatore condiviso nel registry
func (w *Worker) Inc(ctx context.Context, _ *pb.Empty) (*pb.Counter, error) {
	return w.registry.Inc(ctx, &pb.Empty{})
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("uso: go run worker.go <porta> <peso>")
		return
	}

	port := os.Args[1]
	weight, _ := strconv.Atoi(os.Args[2])

	// Connessione al registry
	conn, _ := grpc.Dial("localhost:5000", grpc.WithInsecure())
	defer conn.Close()

	regClient := pb.NewRegistryClient(conn)

	// Registrazione
	regClient.Register(context.Background(), &pb.ServerInfo{
		Host:   "localhost",
		Port:   int32(mustAtoi(port)),
		Weight: int32(weight),
	})

	// Avvio worker
	lis, _ := net.Listen("tcp", ":"+port)
	grpcServer := grpc.NewServer()

	pb.RegisterWorkerServer(grpcServer, &Worker{
		registry: regClient,
	})

	log.Println("[WORKER] running on port", port)
	grpcServer.Serve(lis)
}

func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
