package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cliente struct {
	conexao net.Conn
	nome    string
	jogoID  string
	estado  string
}

type Jogador struct{
	cliente *Cliente
	mao []string
	pontos int
	//skins map[string]*string //TODO: implemantar as esteticas de cartas
}

type Partida struct {
	ID        string
	jogadores map[string]*Jogador
	turno string
}

var clientes = make(map[string]*Cliente)
var partidas = make(map[string]*Partida)
var filaEspera []*Cliente
var clientesMutex = &sync.Mutex{}
var partidasMutex = &sync.Mutex{}
var esperaMutex = &sync.Mutex{}

func main() {
	fmt.Println("Servidor iniciado, aguardando conexões na porta 8080...")
	ouvinte, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("\033[31mErro ao iniciar o servidor:", err, "\033[0m")
		return
	}
	defer ouvinte.Close()

	for {

		conexao, err := ouvinte.Accept()
		if err != nil {
			fmt.Println("\033[31mErro ao aceitar conexão:", err, "\033[0m")
			continue
		}
		go lidarComConexao(conexao)
	}
}

func lidarComConexao(conexao net.Conn) {
	fmt.Println("\033[32mNovo cliente conectado:", conexao.RemoteAddr().String(), "\033[0m")
	conexao.Write([]byte("Bem vindo! Digite seu nome:\n-"))
	leitor := bufio.NewReader(conexao)
	nome, _ := leitor.ReadString('\n')
	nome = strings.TrimSpace(nome)

	cliente := &Cliente{conexao: conexao, nome: nome, estado: ""}

	clientesMutex.Lock()
	clientes[nome] = cliente
	clientesMutex.Unlock()

	cliente.conexao.Write([]byte("Digite 'Procurar' para entrar em uma partida\n-"))

	defer desconectarCliente(cliente)

	for {
		mensagem, err := leitor.ReadString('\n')
		if err == io.EOF {
			fmt.Println("Cliente desconectado:", conexao.RemoteAddr().String())
			//desconectarCliente(cliente)
			break
		}
		if err != nil {
			fmt.Println("\033[31mErro ao ler dados do cliente:", err, "\033[0m")
			//desconectarCliente(cliente)
			break
		}
		mensagem = strings.TrimSpace(mensagem)

		if cliente.jogoID != "" {
			encaminharMensagem(cliente, mensagem)
		} else if strings.ToLower(mensagem) == "procurar" && cliente.estado == "" {
			addFilaEspera(cliente)
		} else {
			switch cliente.estado {
			case "":
				conexao.Write([]byte("Digite 'Procurar' para começar a buscar uma partida:\n-"))
			case "esperando":
				conexao.Write([]byte("Estamos buscando um adversário! espere\n-"))
			}
		}

	}
}

func addFilaEspera(cliente *Cliente) {
	cliente.estado = "esperando"
	esperaMutex.Lock()
	filaEspera = append(filaEspera, cliente)
	esperaMutex.Unlock()
	cliente.conexao.Write([]byte("Buscando adiversário, espere enquanto buscamos um adiversário!\n-"))
	fmt.Printf("Jogador %s entrou na fila de espera\n", cliente.nome)
	verificarEspera()
}

func verificarEspera() {
	esperaMutex.Lock()
	defer esperaMutex.Unlock()

	if len(filaEspera) >= 2 {
		Cliente1 := filaEspera[0]
		Cliente2 := filaEspera[1]

		jogador1 := &Jogador{cliente: Cliente1, mao: make([]string, 0), pontos: 0}
		jogador2 := &Jogador{cliente: Cliente2, mao: make([]string, 0), pontos: 0}


		filaEspera = filaEspera[2:]
		fmt.Printf("INICIANDO PARTIDA: %s x %s\n", Cliente1.nome, Cliente2.nome)
		jogoID := "jogo:" + strconv.FormatInt(time.Now().UnixNano(), 10)

		partidasMutex.Lock()
		partida := &Partida{ID: jogoID, jogadores: make(map[string]*Jogador)}
		primeiroJogador := rand.Intn(2)
		switch primeiroJogador{
		case 0:
			partida.turno=Cliente1.nome
		case 1:
			partida.turno=Cliente2.nome
		}
		partidas[jogoID] = partida
		partida.jogadores[Cliente1.nome] = jogador1
		partida.jogadores[Cliente2.nome] = jogador2
		Cliente1.estado = "jogando"
		Cliente2.estado = "jogando"
		Cliente1.jogoID = jogoID
		Cliente2.jogoID = jogoID
		partidasMutex.Unlock()

		mensagem := fmt.Sprintf("\nAdversário encontrado: %s\n-", Cliente2.nome)
		Cliente1.conexao.Write([]byte(mensagem))
		mensagem = fmt.Sprintf("\nAdversário encontrado: %s\n-", Cliente1.nome)
		Cliente2.conexao.Write([]byte(mensagem))
		switch primeiroJogador{
		case 0:
			Cliente1.conexao.Write([]byte("Você é o primeiro a jogar!\n-"))
			Cliente2.conexao.Write([]byte("Você é o segundo a jogar!\n-"))
		case 1:
			Cliente2.conexao.Write([]byte("\tVocê é o primeiro a jogar!\n-"))
			Cliente1.conexao.Write([]byte("\tVocê é o segundo a jogar!\n-"))
		}
	}
}

func desconectarCliente(cliente *Cliente) {
	esperaMutex.Lock()
	novaFila := []*Cliente{}
	for _, c := range filaEspera {
		if c.nome != cliente.nome {
			novaFila = append(novaFila, c)
		}
	}
	filaEspera = novaFila
	esperaMutex.Unlock()

	if cliente.jogoID != "" {
		partidasMutex.Lock()
		partida, ok := partidas[cliente.jogoID]
		if ok {
			for _, jogador := range partida.jogadores {
				if jogador.cliente.nome != cliente.nome {
					mensagem := fmt.Sprintf("\t------! %s saiu do jogo, a partida acabou------ pressione 'Enter' para voltar ao menu\n-", cliente.nome)
					jogador.cliente.conexao.Write(([]byte(mensagem)))
					jogador.cliente.jogoID = ""
					jogador.cliente.estado = ""
				}
			}
			delete(partidas, cliente.jogoID)
			cliente.jogoID = ""
			cliente.estado = ""
		}
		partidasMutex.Unlock()

	}

	clientesMutex.Lock()
	delete(clientes, cliente.nome)
	clientesMutex.Unlock()

	cliente.conexao.Close()
	fmt.Printf("Cliente desconectado: %s\n", cliente.nome)

}

func encaminharMensagem(remetente *Cliente, mensagem string) {
	partidasMutex.Lock()
	partida, ok := partidas[remetente.jogoID]
	partidasMutex.Unlock()

	if ok {
		if partida.turno == remetente.nome{
			for _, destinatario := range partida.jogadores {
				if destinatario.cliente.nome != remetente.nome {
					mensagem := fmt.Sprintf("[%s]: %s\n-", remetente.nome, mensagem)
					destinatario.cliente.conexao.Write([]byte(mensagem))
					partida.turno = destinatario.cliente.nome
					destinatario.cliente.conexao.Write([]byte(">>> É o seu turno!\n-"))
					remetente.conexao.Write([]byte(">>> Turno do seu oponente\n-"))
				}
			}

		}else{
			remetente.conexao.Write([]byte("Não é seu turno! espere o adverário jogar\n-"))
		}
	}
}
