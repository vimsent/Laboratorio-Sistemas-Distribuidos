package main

import (
	"context"
	"encoding/csv"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync" // Add sync for mutex
	"time"

	// Importamos el código generado por protoc
	pb "lester/proto/grpc/proto" // Reemplaza con el path de tu módulo

	"google.golang.org/grpc"
)

// Definimos una struct para nuestro servidor. Debe embeber el UnimplementedGreeterServer.
// Esto asegura la compatibilidad hacia adelante si se añaden más RPCs al servicio.
type server struct {
	pb.UnimplementedLesterServiceServer
	offers       []*pb.MichaelResponse
	currentIndex int32
	mutex        sync.Mutex // Add mutex for thread safety
}

func NewServer() *server {
	s := &server{
		currentIndex: 0, // Initialize index to 0
	}
	s.loadOffersFromCSV("ofertas_pequeno.csv")
	return s
}

func (s *server) loadOffersFromCSV(filename string) {
	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	reader.Comment = '#'        // Allow comments in CSV
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV: %v", err)
	}

	// Skip header row and parse data
	for i, record := range records {
		if i == 0 { // Skip header
			continue
		}

		if len(record) >= 4 {
			botin, _ := strconv.ParseUint(record[0], 10, 64)
			pFranklin, _ := strconv.ParseInt(record[1], 10, 32)
			pTrevor, _ := strconv.ParseInt(record[2], 10, 32)
			rPolicial, _ := strconv.ParseInt(record[3], 10, 32)

			offer := &pb.MichaelResponse{
				Botin:     botin,
				PFranklin: int32(pFranklin),
				PTrevor:   int32(pTrevor),
				RPolicial: int32(rPolicial),
			}
			s.offers = append(s.offers, offer)
		}
	}

	log.Printf("Loaded %d offers from CSV", len(s.offers))
}

// MichaelOffer es la implementación de la función definida en el archivo .proto.
// Esta es la lógica real que se ejecuta cuando un cliente llama a este RPC.
func (s *server) MichaelOffer(ctx context.Context, in *pb.MichaelRequest) (*pb.MichaelResponse, error) {
	log.Printf("Recibida petición de: %v", in.GetOffer())

	s.mutex.Lock()         // Lock for thread safety
	defer s.mutex.Unlock() // Unlock when function returns
	if in.GetOffer() == "aceptar" {
		log.Printf("Aceptación de oferta recibida")
		log.Printf("Enviando confirmación de fin de fase...")
		end_phase := "end_phase"
		return &pb.MichaelResponse{
			// Assuming your pb.MichaelResponse has an appropriate field
			// For example, if it has a field called "Value":
			POferta: end_phase,
		}, nil
	}

	rand.Seed(time.Now().UnixNano())
	if rand.Float64() > 0.9 {
		log.Printf("No hay ofertas disponibles para Michael")
		log.Printf("Enviando mensaje a Michael...")
		reject := "rechazo"
		return &pb.MichaelResponse{
			// Assuming your pb.MichaelResponse has an appropriate field
			// For example, if it has a field called "Value":
			POferta: reject,
		}, nil
	}

	// Get current offer and move to next one
	if s.currentIndex == 3 {
		log.Printf("3 rechazos detectados, se esperan 10 segundos...")
		time.Sleep(10 * time.Second)
	}
	offer := s.offers[s.currentIndex]
	s.currentIndex = (s.currentIndex + 1) % int32(len(s.offers))

	log.Printf("Returning offer %d of %d", s.currentIndex, len(s.offers))
	return offer, nil
	/*
		test := int32(99) // Use int32 to match protobuf types
		return &pb.MichaelResponse{
			// Assuming your pb.MichaelResponse has an appropriate field
			// For example, if it has a field called "Value":
			RPolicial: test,
		}, nil
	*/

}

func main() {
	// Create the server with loaded CSV data
	s := NewServer()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLesterServiceServer(grpcServer, s)

	log.Printf("Server starting on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
