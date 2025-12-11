package main

import (
	"context"
	"log"
	"time"

	pb "michael/proto/grpc/proto"

	"google.golang.org/grpc"
)

const (
	address_lester   = "10.35.168.89:50051"
	address_trevor   = "10.35.168.90:50052"
	address_franklin = "10.35.168.112:50053"
)

// Function to communicate with either Trevor or Franklin
func communicateWithDistractionService(ctx context.Context, address string, message string, exito int32, isTrevor bool) (string, error) {
	// Connect to the distraction service
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if isTrevor {
		// Communicate with Trevor
		client := pb.NewTrevorServiceClient(conn)
		response, err := client.Distraccion(ctx, &pb.TrevorRequest{Notificar: message, PTrevor: exito})
		if err != nil {
			return "", err
		}
		return response.Resultado, nil
	} else {
		// Communicate with Franklin
		client := pb.NewFranklinServiceClient(conn)
		response, err := client.Distraccion(ctx, &pb.FranklinRequest{Notificar: message})
		if err != nil {
			return "", err
		}
		return response.Resultado, nil
	}
}

func main() {
	//Esperar 2 segundos para iniciar a Michael
	//Para asegurarse de que el servidor este listo antes de que el cliente intente conectarse
	log.Println("Esperando 2 segundos para iniciar el programa...")
	time.Sleep(2 * time.Second)

	//Conectar al servidor gRPC
	conn, err := grpc.Dial(address_lester, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor: %v", err)
	}
	defer conn.Close()

	//Crear un nuevo cliente gRPC
	//pb es el paquete generado por el compilador de gRPC a partir del archivo proto
	client := pb.NewLesterServiceClient(conn)

	// 3. Preparamos el contexto y los datos para la llamada remota.
	//    Un contexto puede llevar deadlines, cancelaciones, y otros valores a través
	//    de las llamadas.
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	// Use background context without timeout
	ctx := context.Background()
	//------------------------------------FASE 1------------------------------------
	operation := "solicitar_trabajo"
	log.Println("Inicio FASE 1")
	log.Println("Michael solicita trabajo a Lester...")
	count := 0
	var Botin_job uint64
	var pFranklin_job int32
	var pTrevor_job int32
	var rPolicial_job int32
	for {
		count++
		log.Printf("Intento %d", count)
		resp, err := client.MichaelOffer(ctx, &pb.MichaelRequest{Offer: operation})
		if err != nil {
			log.Fatalf("Error al enviar la operación: %v", err)
		}
		//Imprimir el resultado de la operación
		//El servidor gRPC devuelve un mensaje de tipo OperationResponse, que contiene el resultado de la operación
		if resp.POferta == "rechazo" {
			log.Printf("Lester no tiene oferta (10%% probabilidad)")
			log.Printf("Se vuelve a solicitar trabajo...")
			continue
		} else {
			log.Printf("Oferta de Lester:")
			log.Printf("Botín: %d\n", resp.Botin)
			log.Printf("Éxito Franklin: %d%%", resp.PFranklin)
			log.Printf("Éxito Trevor: %d%%", resp.PTrevor)
			log.Printf("Riesgo Policial: %d%%", resp.RPolicial)
			if resp.PFranklin > 50 || resp.PTrevor > 50 {
				if resp.RPolicial < 80 {
					log.Printf("Trabajo aceptado")
					end_phase := "aceptar"
					Botin_job = resp.Botin
					pFranklin_job = resp.PFranklin
					pTrevor_job = resp.PTrevor
					rPolicial_job = resp.RPolicial
					resp2, err := client.MichaelOffer(ctx, &pb.MichaelRequest{Offer: end_phase})
					if err != nil {
						log.Fatalf("Error al enviar la operación: %v", err)
					}
					if resp2.POferta == "end_phase" {
						log.Printf("Lester confirma fin de fase 1")
						break
					}
				}
			} else {
				log.Printf("Trabajo no cumple requisitos mínimos")
				log.Printf("Consultando a Lester por nueva oferta...")
			}
			time.Sleep(2 * time.Second)
		}
	}
	//--------------------------------------FIN-FASE-1-----------------------------------
	log.Printf("%d %d %d %d", Botin_job, pFranklin_job, pTrevor_job, rPolicial_job)

	//----------------------------------------FASE-2-------------------------------------
	log.Println("Inicio FASE 2")

	// Elegir en base a mayor probabilidad
	var chosenPartner string
	var useTrevor bool
	var exito int32
	var partnerAddress string

	if pTrevor_job > pFranklin_job {
		chosenPartner = "Trevor"
		useTrevor = true
		exito = pTrevor_job
		partnerAddress = address_trevor
	} else {
		chosenPartner = "Franklin"
		useTrevor = false
		partnerAddress = address_franklin
	}

	log.Printf("Contactando a %s (mayor probabilidad de éxito)", chosenPartner)

	// Use the reusable function to communicate with the chosen partner
	message := "iniciar distracción"
	result, err := communicateWithDistractionService(ctx, partnerAddress, message, exito, useTrevor)
	if err != nil {
		log.Fatalf("Error al comunicarse con %s: %v", chosenPartner, err)
	}

	log.Printf("Respuesta de %s: %s", chosenPartner, result)
}
