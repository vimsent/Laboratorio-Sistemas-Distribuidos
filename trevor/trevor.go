package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"time"

	pb "trevor/proto/grpc/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTrevorServiceServer
}

func (s *server) Distraccion(ctx context.Context, in *pb.TrevorRequest) (*pb.TrevorResponse, error) {
	log.Printf("Recibida petición de: %v", in.GetNotificar())
	turnosNecesarios := 200 - in.GetPTrevor()
	rand.Seed(time.Now().UnixNano())
	log.Printf("Turnos necesarios: %d", turnosNecesarios)
	turnos := int32(0)
	imprevisto := turnosNecesarios / 2
	for turnos < turnosNecesarios {
		if turnos == imprevisto {
			if rand.Float64() > 0.9 {
				log.Printf("Trevor está borracho, el trabajo ha fracasado...")
				return &pb.TrevorResponse{
					Resultado: "fracaso",
				}, nil
			}
		}
		log.Printf("Creando distracción...")
		log.Printf("Turno %d", turnos)
		turnos++
	}
	log.Printf("Trevor termina distracción con éxito")
	log.Printf("Enviando resultado a Michael...")
	return &pb.TrevorResponse{
		Resultado: "exito",
	}, nil

}

func main() {
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTrevorServiceServer(grpcServer, &server{})

	log.Printf("Server starting on port 50052...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
