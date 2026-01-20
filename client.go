package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	pb "grpc-project/proto"

	"google.golang.org/grpc"
)

type Client struct {
	servers []*pb.ServerInfo
	index   int
	cache   map[string]string
}

func (c *Client) lookup() {
	conn, _ := grpc.Dial("localhost:5000", grpc.WithInsecure())
	defer conn.Close()

	reg := pb.NewRegistryClient(conn)

	list, _ := reg.GetServers(context.Background(), &pb.Empty{})
	c.servers = list.Servers

	fmt.Println("\n[CLIENT] Servers discovered:")
	for _, s := range c.servers {
		fmt.Printf(" - %s:%d (w=%d)\n", s.Host, s.Port, s.Weight)
	}
}

// ===== LOAD BALANCING =====

func (c *Client) nextStateless() *pb.ServerInfo {

	if len(c.servers) == 0 {
		fmt.Println("[ERRORE] Nessun server disponibile. Fai lookup (opzione 4).")
		return nil
	}

	s := c.servers[c.index%len(c.servers)]
	c.index++
	return s
}

func (c *Client) nextStateful() *pb.ServerInfo {

	if len(c.servers) == 0 {
		fmt.Println("[ERRORE] Nessun server disponibile. Fai lookup (opzione 4).")
		return nil
	}

	total := 0
	for _, s := range c.servers {
		total += int(s.Weight)
	}

	if total == 0 {
		return c.servers[0]
	}

	r := rand.Intn(total)
	sum := 0

	for _, s := range c.servers {
		sum += int(s.Weight)
		if r < sum {
			return s
		}
	}

	return c.servers[0]
}

// ===== CHIAMATE =====

func (c *Client) callEcho(msg string, stateful bool) string {

	if v, ok := c.cache[msg]; ok {
		fmt.Println("[CACHE HIT]")
		return v
	}

	var s *pb.ServerInfo
	strategy := "STATELESS (Round Robin)"

	if s == nil {
		return "[nessun server]"
	}

	if stateful {
		s = c.nextStateful()
		strategy = "STATEFUL (Weighted)"
	} else {
		s = c.nextStateless()
	}

	fmt.Println("----------------------------------")
	fmt.Printf("[CLIENT] Strategia: %s\n", strategy)
	fmt.Printf("[CLIENT] Server scelto: %s:%d\n", s.Host, s.Port)
	fmt.Printf("[CLIENT] Peso server: %d\n", s.Weight)
	fmt.Println("----------------------------------")

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	conn, _ := grpc.Dial(addr, grpc.WithInsecure())
	defer conn.Close()

	worker := pb.NewWorkerClient(conn)

	res, _ := worker.Echo(context.Background(), &pb.Text{Msg: msg})

	c.cache[msg] = res.Msg
	return res.Msg
}

func (c *Client) callAdd(a, b int32) int32 {
	s := c.nextStateless()

	fmt.Println("----------------------------------")
	fmt.Printf("[CLIENT] Add su server: %s:%d (peso=%d)\n",
		s.Host, s.Port, s.Weight)
	fmt.Println("----------------------------------")

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	conn, _ := grpc.Dial(addr, grpc.WithInsecure())
	defer conn.Close()

	worker := pb.NewWorkerClient(conn)

	res, _ := worker.Add(context.Background(), &pb.Numbers{A: a, B: b})
	return res.Value
}

// 🔴 CHIAMATA STATEFUL – CONTATORE CONDIVISO
func (c *Client) callInc() int32 {

	// non importa quale worker: tutti chiamano il registry
	s := c.nextStateful()

	fmt.Println("----------------------------------")
	fmt.Printf("[CLIENT] Incremento contatore tramite worker: %s:%d\n",
		s.Host, s.Port)
	fmt.Printf("[CLIENT] (lo stato è condiviso nel registry)\n")
	fmt.Println("----------------------------------")

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	conn, _ := grpc.Dial(addr, grpc.WithInsecure())
	defer conn.Close()

	worker := pb.NewWorkerClient(conn)

	res, _ := worker.Inc(context.Background(), &pb.Empty{})
	return res.Value
}

// ===== MENU =====

func menu() {
	fmt.Println("\n===== CLIENT MENU =====")
	fmt.Println("1 - Echo (stateless)")
	fmt.Println("2 - Add (stateless)")
	fmt.Println("3 - Inc (STATEFUL condiviso)")
	fmt.Println("4 - Ricarica lista server")
	fmt.Println("0 - Esci")
	fmt.Print("> ")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	client := &Client{
		cache: make(map[string]string),
	}

	client.lookup()

	reader := bufio.NewReader(os.Stdin)

	for {
		menu()

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {

		case "1":
			fmt.Print("Testo: ")
			msg, _ := reader.ReadString('\n')
			msg = strings.TrimSpace(msg)

			fmt.Println("Risposta:", client.callEcho(msg, false))

		case "2":
			var a, b int32
			fmt.Print("a: ")
			fmt.Scan(&a)
			fmt.Print("b: ")
			fmt.Scan(&b)

			fmt.Println("Risultato:", client.callAdd(a, b))

		case "3":
			v := client.callInc()
			fmt.Println("[COUNTER CONDIVISO] valore =", v)

		case "4":
			client.lookup()

		case "0":
			log.Println("Uscita client")
			return

		default:
			fmt.Println("Scelta non valida")
		}
	}
}
