# Nome do binário (exigido pela AWS quando usamos runtime "provided.al2")
BINARY_NAME=bootstrap
ZIP_NAME=lambda.zip
GOARCH=amd64
GOOS=linux

# Caminho de saída
BUILD_DIR=.
ZIP_PATH=../infra/$(ZIP_NAME)

.PHONY: all build zip clean

# Tarefa padrão
all: build zip

# Compila o código Go para Linux
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Empacota o binário como lambda.zip
zip: build
	zip -j $(ZIP_NAME) $(BINARY_NAME)
	mv $(ZIP_NAME) $(ZIP_PATH)

# Limpa arquivos gerados
clean:
	rm -f $(BINARY_NAME) $(ZIP_PATH)
