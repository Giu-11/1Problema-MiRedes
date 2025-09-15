# 🃏 Projeto de Jogo de Cartas em Go

Este é um projeto de um jogo de cartas multiplayer (estilo 21/Blackjack simplificado) desenvolvido em Go. A arquitetura é baseada em cliente-servidor, permitindo que múltiplos jogadores se conectem a um servidor central, encontrem oponentes e joguem em tempo real.

## 📋 Pré-requisitos

Para executar o projeto, você precisará ter as seguintes ferramentas instaladas:

* [Go](https://go.dev/doc/install) (versão 1.25.0)  
* [Docker](https://docs.docker.com/get-docker/)  
* [Docker Compose](https://docs.docker.com/compose/install/)  

---

## 🚀 Como Executar o Projeto

Existem quatro maneiras principais de executar a aplicação, dependendo das suas necessidades.

### Método 1: Localmente (Sem Docker)

1. **Iniciar o Servidor:**
   ```
   cd servidor
   go run .
   ```

2. **Iniciar o Cliente:**
   ```
   cd cliente
   go run .
   ```
   Você pode iniciar quantos clientes quiser, cada um em seu próprio terminal.

   > 💡 **Dica:** Se o servidor estiver rodando em outra máquina na mesma rede, você pode conectar o cliente a ele especificando o endereço como argumento:  
   > ```
   > go run . <endereço_ip_do_servidor>:8080
   > ```

---

### Método 2: Apenas o Servidor com Docker

1. **Construir a Imagem Docker do Servidor:**
   ```
   docker build -t servidor-jogo-cartas .
   ```

2. **Rodar o Contêiner do Servidor:**
   ```
   docker run --rm -p 8080:8080 servidor-jogo-cartas
   ```

3. **Conectar Clientes Locais:**
   Execute os clientes localmente como no **Método 1**.

---

### Método 3: Cliente e Servidor com Docker Compose

1. **Construir as Imagens de Ambos os Serviços:**
   ```
   docker-compose build
   ```

2. **Iniciar os Serviços (usando dois terminais):**

   * **Servidor:**
     ```
     docker-compose up servidor
     ```

   * **Cliente:**
     ```
     docker-compose run --rm cliente
     ```

3. **Encerrar Tudo:**
   ```
   docker-compose down
   ```

---

### Método 4: Cliente e Servidor em Docker (Máquinas Diferentes)

Este método é útil quando o **servidor** e o **cliente** vão rodar em hosts diferentes da mesma rede local.

1. **Rodar o Servidor (em uma máquina):**  
   Você pode escolher entre **Docker Compose** ou rodar o servidor sozinho:

   *Com Docker Compose:*
   ```
   docker-compose up servidor
   ```

   *Ou apenas com Docker direto:*
   ```
   docker build -t servidor-jogo-cartas .
   docker run --rm -p 8080:8080 servidor-jogo-cartas
   ```

   O servidor ficará exposto na porta `8080` do host.  
   > 💡 Para descobrir o IP da máquina do servidor (Linux), use:  
   > ```
   > hostname -I
   > ```

2. **Rodar o Cliente (em outra máquina):**  
   Primeiro faça o build:
   ```
   docker build -t cliente-jogo-cartas -f cliente/Dockerfile .
   ```

   Depois execute, apontando para o IP da máquina do servidor:
   ```
   docker run --rm -it cliente-jogo-cartas ./cliente <endereço_ip_do_servidor>:8080
   ```

   Exemplo:
   ```
   docker run --rm -it cliente-jogo-cartas ./cliente 192.168.0.42:8080
   ```

---

## 🧪 Rodando os Testes

O projeto inclui testes de concorrência para validar a lógica de matchmaking e a distribuição de cartas do servidor.

Para executá-los, navegue até a pasta raiz e use o comando padrão de testes do Go:

```
cd servidor
go test -v
```

* A flag `-v` (verbose) exibe os resultados detalhados de cada teste.  
