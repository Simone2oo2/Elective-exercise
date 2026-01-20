package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "grpc-project/proto"

	"google.golang.org/grpc"
)

type RegistryServer struct {
	pb.UnimplementedRegistryServer

	mu      sync.Mutex
	servers []*pb.ServerInfo
	counter int32
}

func (r *RegistryServer) Register(ctx context.Context, info *pb.ServerInfo) (*pb.ServerId, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.servers = append(r.servers, info)

	id := fmt.Sprintf("%s:%d", info.Host, info.Port)

	// 🔴 LOG PIÙ DETTAGLIATO
	log.Println("======================================")
	log.Println("[REGISTRY] NUOVO WORKER REGISTRATO")
	log.Printf(" - Indirizzo : %s\n", id)
	log.Printf(" - Peso      : %d\n", info.Weight)
	log.Printf(" - Totale worker attivi: %d\n", len(r.servers))
	log.Println("======================================")

	return &pb.ServerId{Id: id}, nil
}

func (r *RegistryServer) GetServers(ctx context.Context, _ *pb.Empty) (*pb.ServerList, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return &pb.ServerList{Servers: r.servers}, nil
}

// CONTATORE CONDIVISO (opzione B)
func (r *RegistryServer) Inc(ctx context.Context, _ *pb.Empty) (*pb.Counter, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	return &pb.Counter{Value: r.counter}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRegistryServer(grpcServer, &RegistryServer{})

	log.Println("[REGISTRY] running on port 5000")
	grpcServer.Serve(lis)
}
