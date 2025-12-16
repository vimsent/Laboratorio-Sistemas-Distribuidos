# GTA Heist Simulator - Sistema Distribuido

## Descripción General

Este proyecto simula un atraco bancario inspirado en Grand Theft Auto V, implementado como un sistema distribuido utilizando **gRPC**, **RabbitMQ** y **Docker**. El sistema coordina múltiples servicios que representan a los personajes del juego (Michael, Lester, Trevor y Franklin) trabajando juntos para completar un robo exitoso.

## Arquitectura del Sistema

```
┌─────────────────────────────────────────────────────────────┐
│                         MICHAEL                             │
│                    (Coordinador/Cliente)                    │
└────────────┬────────────────────────────────────┬───────────┘
             │                                    │
             ├─── gRPC ───────┐                   │
             │                │                   │
             ▼                ▼                   ▼
      ┌──────────┐      ┌─────────┐       ┌──────────┐
      │  LESTER  │      │ TREVOR  │       │ FRANKLIN │
      │ (Planif.)│      │ (Agente)│       │ (Agente) │
      └────┬─────┘      └────┬────┘       └────┬─────┘
           │                 │                  │
           └────────┬────────┴──────────────────┘
                    │
                    ▼
            ┌────────────────┐
            │   RABBITMQ     │
            │ (Notificaciones│
            │  de Estrellas) │
            └────────────────┘
```

## Fases del Atraco

### **Fase 1: Negociación** 
**Responsable:** Lester (Servidor) ↔ Michael (Cliente)

Michael solicita trabajos disponibles a Lester, quien responde con ofertas de atracos que contienen:
- **Botín base** (`uint64`): Cantidad de dinero a robar
- **Probabilidad de éxito de Franklin** (`int32`): 0-100%
- **Probabilidad de éxito de Trevor** (`int32`): 0-100%
- **Riesgo policial** (`int32`): Nivel de alerta policial

**Criterios de aceptación:**
- Al menos un personaje debe tener >50% de probabilidad de éxito
- Riesgo policial debe ser <80%

**Características especiales:**
- 10% de probabilidad de que Lester no tenga ofertas disponibles
- Después de 3 rechazos, hay una penalización de 10 segundos
- Las ofertas se cargan desde un archivo CSV

### **Fase 2: Distracción** 
**Responsable:** Franklin o Trevor (el de mayor probabilidad)

El personaje elegido crea una distracción para facilitar el robo:
- **Duración:** 200 - probabilidad_éxito turnos (cada turno = 50ms)
- **Imprevisto:** A la mitad del trabajo hay una probabilidad de fracaso:
  - **Trevor:** 10% de estar borracho y fallar
  - **Franklin:** 10% de que Chop (su perro) ladre y alerte

**Resultado:** "exito" o "fracaso"

### **Fase 3: El Golpe** 
**Responsable:** El personaje que NO hizo la distracción

Esta es la fase principal del atraco, donde ocurre lo siguiente:

1. **Lester inicia notificaciones de estrellas** vía RabbitMQ
2. El personaje ejecuta el robo durante 200 - probabilidad_éxito turnos
3. **Sistema de estrellas:**
   - Frecuencia: cada (100 - riesgo_policial) * 100ms
   - Se incrementa de 1 en 1 hasta alcanzar el límite

**Habilidades Especiales:**

#### Franklin + Chop
- **Activación:** Al alcanzar 3 estrellas
- **Efecto:** Gana $1,000 extra por turno restante
- **Límite de fracaso:** 5 estrellas

#### Trevor + Furia
- **Activación:** Al alcanzar 5 estrellas
- **Efecto:** Aumenta el límite de fracaso de 5 a 7 estrellas
- **Botín extra:** Ninguno

**Resultado:** Objeto `GolpeResponse` con:
- `exito` (bool)
- `motivoFallo` (string)
- `botinExtra` (int64)
- `estrellasFinales` (int32)

### **Fase 4: Reparto del Botín** 
**(TODO - No implementada)**

Distribución planificada del botín entre los participantes.

## Tecnologías Utilizadas

### **Lenguaje de Programación**
- **Go 1.24.2**: Lenguaje principal del proyecto

### **Comunicación entre Servicios**
- **gRPC**: Comunicación síncrona entre servicios
  - Llamadas RPC tipadas y eficientes
  - Definición de servicios con Protocol Buffers
- **Protocol Buffers v3**: Serialización de datos
  - Definición de mensajes y servicios en `msg.proto`

### **Sistema de Mensajería**
- **RabbitMQ 3 Management**: Cola de mensajes asíncrona
  - Notificaciones de estrellas policiales
  - Patrón Publisher/Subscriber
  - Persistencia de mensajes

### **Librerías de Go**
- `google.golang.org/grpc`: Cliente y servidor gRPC
- `google.golang.org/protobuf`: Soporte para Protocol Buffers
- `github.com/streadway/amqp`: Cliente RabbitMQ

### **Infraestructura**
- **Docker**: Contenerización de servicios
- **Docker Compose**: Orquestación de contenedores
  - Red compartida `labnet`
  - Healthchecks para dependencias
  - Variables de entorno para configuración

## Estructura del Proyecto

```
.
├── docker-compose.yml          # Orquestación de servicios
├── README.md                   # Este archivo
│
├── lester/                     # Servicio de planificación
│   ├── Dockerfile
│   ├── Makefile
│   ├── lester.go              # Lógica del servidor
│   ├── ofertas_pequeno.csv    # Base de datos de ofertas
│   ├── go.mod / go.sum
│   └── proto/
│       ├── msg.proto          # Definición de servicios
│       └── grpc/proto/        # Código generado
│
├── trevor/                     # Servicio del agente Trevor
│   ├── Dockerfile
│   ├── Makefile
│   ├── trevor.go              # Lógica del servidor
│   ├── go.mod / go.sum
│   └── proto/
│       ├── msg.proto
│       └── grpc/proto/
│
├── franklin/                   # Servicio del agente Franklin
│   ├── Dockerfile
│   ├── Makefile
│   ├── franklin.go            # Lógica del servidor
│   ├── go.mod / go.sum
│   └── proto/
│       ├── msg.proto
│       └── grpc/proto/
│
└── michael/                    # Servicio coordinador (cliente)
    ├── Dockerfile
    ├── Makefile
    ├── michael.go             # Lógica del cliente
    ├── go.mod / go.sum
    └── proto/
        ├── msg.proto
        └── grpc/proto/
```

## Instalación y Ejecución

### Prerequisitos
- Docker y Docker Compose instalados
- Go 1.24.2+ (solo para desarrollo local)
- `protoc` (Protocol Buffers compiler) si se modifican los `.proto`

### Compilación y Ejecución

1. **Clonar el repositorio**
```bash
git clone <repository-url>
cd gta-heist-simulator
```

2. **Construir y ejecutar con Docker Compose**
```bash
docker-compose up --build
```

Esto iniciará todos los servicios:
- RabbitMQ en puertos 5672 (AMQP) y 15672 (Management UI)
- Lester en puerto 50051
- Trevor en puerto 50052
- Franklin en puerto 50053
- Michael (coordinador)

3. **Ver logs**
```bash
docker-compose logs -f michael    # Logs del coordinador
docker-compose logs -f lester     # Logs de Lester
docker-compose logs -f trevor     # Logs de Trevor
docker-compose logs -f franklin   # Logs de Franklin
```

4. **Detener servicios**
```bash
docker-compose down
```

### Regenerar Protocol Buffers (Desarrollo)

Si modificas los archivos `.proto`:

```bash
cd lester  # o trevor, franklin, michael
protoc --go_out=. --go-grpc_out=. msg.proto
```

## Configuración

### Variables de Entorno

Cada servicio puede configurarse con las siguientes variables:

```yaml
RABBITMQ_HOST: "rabbitmq"          # Host de RabbitMQ
RABBITMQ_PORT: "5672"              # Puerto de RabbitMQ
RABBITMQ_DEFAULT_USER: "user"      # Usuario de RabbitMQ
RABBITMQ_DEFAULT_PASS: "pass"      # Contraseña de RabbitMQ
```

### Archivo de Ofertas

Las ofertas se definen en `lester/ofertas_pequeno.csv`:

```csv
botin_inicial,prob_exito_franklin,prob_exito_trevor,riesgo_policial
1757257,41,43,87
1215098,41,83,
...
```

**Nota:** Los valores vacíos se manejan como 0 por defecto.

##  Flujo de Datos

### 1. Comunicación gRPC (Síncrona)

```
Michael → Lester:  MichaelRequest("solicitar_trabajo")
Lester → Michael:  MichaelResponse{botin, pFranklin, pTrevor, rPolicial}

Michael → Franklin: FranklinRequest("iniciar distracción", pFranklin)
Franklin → Michael: FranklinResponse("exito")

Michael → Trevor:   GolpeRequest("trevor", probabilidad, riesgo)
Trevor → Michael:   GolpeResponse{exito: true, botinExtra: 0}
```

### 2. Comunicación RabbitMQ (Asíncrona)

```json
Lester → Queue(franklin_stars):
{
  "stars": 3,
  "character": "franklin",
  "timestamp": 1234567890
}

Queue → Franklin: Recibe y procesa notificación
```

##  Mensajes de Protocol Buffers

### Servicios Principales

```protobuf
service LesterService {
  rpc MichaelOffer(MichaelRequest) returns (MichaelResponse);
  rpc IniciarNotificaciones(NotificacionRequest) returns (NotificacionResponse);
  rpc DetenerNotificaciones(DetenerRequest) returns (DetenerResponse);
}

service TrevorService {
  rpc Distraccion(TrevorRequest) returns (TrevorResponse);
  rpc IniciarGolpe(GolpeRequest) returns (GolpeResponse);
  rpc ConsultarEstrellas(EstrellasRequest) returns (EstrellasResponse);
  rpc ObtenerBotin(BotinRequest) returns (BotinResponse);
}

service FranklinService {
  rpc Distraccion(FranklinRequest) returns (FranklinResponse);
  rpc IniciarGolpe(GolpeRequest) returns (GolpeResponse);
  rpc ConsultarEstrellas(EstrellasRequest) returns (EstrellasResponse);
  rpc ObtenerBotin(BotinRequest) returns (BotinResponse);
}
```

##  Reporte Final

Al completar (exitosa o fallida) la misión, Michael genera un archivo `Reporte.txt` con:

**En caso de éxito:**
- ID de la misión
- Personajes participantes en cada fase
- Botín base y extra
- Botín total acumulado

**En caso de fracaso:**
- Fase donde ocurrió el fracaso
- Personaje responsable
- Motivo del fracaso
- Botín perdido

##  Problemas Conocidos

### Configuración de Red en Docker Compose

**Problema actual:** Los servicios están configurados con IPs hardcodeadas en lugar de nombres de servicio Docker:

```go
//  Incorrecto (actual)
const address_franklin = "10.35.168.112:50053"

//  Correcto (debería ser)
const address_franklin = "franklin:50053"
```

**Solución:** Modificar las direcciones en `michael/michael.go`:

```go
const (
    address_lester   = "lester:50051"
    address_trevor   = "trevor:50052"
    address_franklin = "franklin:50053"
)
```

Y en `docker-compose.yml` para Franklin:

```yaml
franklin:
  environment:
    - RABBITMQ_HOST=rabbitmq  # En lugar de localhost
```


