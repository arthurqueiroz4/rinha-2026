# Stage 1: Preprocessamento de dados
FROM golang:1.25-alpine AS preprocessor
WORKDIR /app

# Copia go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copia código necessário para preprocessamento
COPY model/ model/
COPY quantization/ quantization/
COPY vptree/ vptree/
COPY normalize/ normalize/
COPY preprocess/ preprocess/

# Copia resources brutos
COPY resources/references.json.gz resources/

# Roda preprocessamento para gerar vptree.bin
RUN go run preprocess/preprocess.go

# Stage 2: Builder da aplicação
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copia go.mod e go.sum primeiro (para cache de dependências)
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o código
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-s -w" -o api .

# Stage 3: Runtime final
FROM alpine:latest
RUN apk add --no-cache curl
COPY --from=builder /app/api /api

# Copia resources do stage de preprocessamento (vptree.bin gerado)
COPY --from=preprocessor /app/resources/vptree.bin /resources/vptree.bin

# Copia outros resources do contexto original
COPY resources/normalization.json /resources/normalization.json
COPY resources/mcc_risk.json /resources/mcc_risk.json

EXPOSE 8080
CMD ["/api"]
