package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"projeto-rede/cartasUtils"
	"projeto-rede/estilo"
	"projeto-rede/protocolo"
	"projeto-rede/servUtils"
	"strconv"
	"sync"
	"time"
)

var clientes = make(map[string]*servUtils.Cliente)
var partidas = make(map[string]*servUtils.Partida)
var filaEspera []*servUtils.Cliente
var clientesMutex = &sync.Mutex{}
var partidasMutex = &sync.Mutex{}
var esperaMutex = &sync.Mutex{}

func main() {
	fmt.Println("Servidor iniciado, aguardando conexões na porta 8080...")
	ouvinte, err := net.Listen("tcp", ":8080")
	if err != nil {
		msg:=fmt.Sprintf("Erro ao iniciar o servidor:%s\n",err)
		estilo.PrintVerm(msg)
		return
	}
	defer ouvinte.Close()

	for {

		conexao, err := ouvinte.Accept()
		if err != nil {
			msg:=fmt.Sprintf("Erro ao aceitar conexão:%s\n",err)
			estilo.PrintVerm(msg)
			continue
		}
		go lidarComConexao(conexao)
	}
}

func lidarComConexao(conexao net.Conn) {
	msg:=fmt.Sprintf("Novo cliente conectado:%s\n", conexao.RemoteAddr().String())
	estilo.PrintVerd(msg)
	decodificador := json.NewDecoder(conexao)
	codificador := json.NewEncoder(conexao)

	cliente := &servUtils.Cliente{Conexao: conexao, Nome: "", Estado: "login", JogoID: ""}

	defer desconectarCliente(cliente)

	sair := false
	for !sair {
		//
		var envelope protocolo.Envelope
		err := decodificador.Decode(&envelope)
		if err == nil {
			switch envelope.Requisicao {
			case "login":
				if cliente.Estado == "login" {
					var dadosLogin protocolo.Login
					err := json.Unmarshal(envelope.Dados, &dadosLogin)
					if err == nil {
						if tentarLogin(dadosLogin) {
							cliente.Nome = dadosLogin.Nome
							cliente.Estado = ""
							addListaClientes(cliente)
							servUtils.EnviarResposta(*codificador, "confirmacao", "login", true)
						} else {
							servUtils.EnviarResposta(*codificador, "confirmacao", "login", false)
						}
					}
				} else {
					servUtils.EnviarAviso(*codificador, "Ação inválida")
				}
			case "procurar":
				if cliente.Estado == "" {
					cliente.Estado = "esperando"
					addFilaEspera(cliente)
					//TODO: função de sair da fila de espera
				} else {
					servUtils.EnviarAviso(*codificador, "Ação inválida")
				}
			case "jogada":
				if cliente.Estado == "jogando" {
					var jogada protocolo.Jogada
					err := json.Unmarshal(envelope.Dados, &jogada)
					if err == nil {
						lidarComJogada(cliente, jogada, *codificador)
					}
				}
			}
		} else {
			fmt.Printf("Cliente %s perdeu a conexão. Erro: %v\n", cliente.Nome, err)
			sair = true
		}
	}
}

func tentarLogin(dadosLogin protocolo.Login) bool {
	if dadosLogin.Nome != "" {
		clientesMutex.Lock()
		defer clientesMutex.Unlock()
		_, existe := clientes[dadosLogin.Nome]
		if !existe{
			return true
		}
	}
	return false
	
}

func addListaClientes(cliente *servUtils.Cliente) {
	clientesMutex.Lock()
	clientes[cliente.Nome] = cliente
	clientesMutex.Unlock()
}

func addFilaEspera(cliente *servUtils.Cliente) {
	cliente.Estado = "esperando"
	esperaMutex.Lock()
	filaEspera = append(filaEspera, cliente)
	esperaMutex.Unlock()
	codificador := json.NewEncoder(cliente.Conexao)
	servUtils.EnviarAviso(*codificador, "Buscando adversário, aguarde...")
	fmt.Printf("Jogador %s entrou na fila de espera\n", cliente.Nome)
	verificarEspera()
}

func verificarEspera() {
	esperaMutex.Lock()
	defer esperaMutex.Unlock()

	if len(filaEspera) >= 2 {
		Cliente1 := filaEspera[0]
		Cliente2 := filaEspera[1]

		codificador1 := json.NewEncoder(Cliente1.Conexao)
		codificador2 := json.NewEncoder(Cliente2.Conexao)

		jogador1 := &servUtils.Jogador{Cliente: Cliente1, Mao: make([]string, 0), Pontos: 0, ParouCartas: false}
		jogador2 := &servUtils.Jogador{Cliente: Cliente2, Mao: make([]string, 0), Pontos: 0, ParouCartas: false}
		Cliente1.Jogador = jogador1
		Cliente2.Jogador = jogador2

		filaEspera = filaEspera[2:]
		msg := fmt.Sprintf("INICIANDO PARTIDA: %s x %s\n", Cliente1.Nome, Cliente2.Nome)
		estilo.PrintCian(msg)
		jogoID := "jogo:" + strconv.FormatInt(time.Now().UnixNano(), 10)

		partidasMutex.Lock()
		partida := &servUtils.Partida{ID: jogoID, Jogadores: make(map[string]*servUtils.Jogador)}
		primeiroJogador := rand.Intn(2)
		switch primeiroJogador {
		case 0:
			partida.Turno = Cliente1.Nome
		case 1:
			partida.Turno = Cliente2.Nome
		}
		clientesMutex.Lock()
		partidas[jogoID] = partida
		partida.Jogadores[Cliente1.Nome] = jogador1
		partida.Jogadores[Cliente2.Nome] = jogador2
		Cliente1.Estado = "jogando"
		Cliente2.Estado = "jogando"
		Cliente1.JogoID = jogoID
		Cliente2.JogoID = jogoID
		clientesMutex.Unlock()
		partida.Cartas = cartasUtils.GeradorCartasEmbaralhadas()
		partidasMutex.Unlock()

		servUtils.EnviarInicioPartida(*codificador1, Cliente2.Nome, partida.Turno)
		servUtils.EnviarInicioPartida(*codificador2, Cliente1.Nome, partida.Turno)

		switch primeiroJogador {
		case 0:
			servUtils.EnviarAviso(*codificador1, "Você é o primeiro a jogar!")
			servUtils.EnviarAviso(*codificador2, "Você é o segundo a jogar!")
		case 1:
			servUtils.EnviarAviso(*codificador2, "Você é o primeiro a jogar!")
			servUtils.EnviarAviso(*codificador1, "Você é o segundo a jogar!")
		}
	}
}



func desconectarCliente(cliente *servUtils.Cliente) {
	esperaMutex.Lock()
	novaFila := []*servUtils.Cliente{}
	for _, c := range filaEspera {
		if c.Nome != cliente.Nome {
			novaFila = append(novaFila, c)
		}
	}
	filaEspera = novaFila
	esperaMutex.Unlock()

	if cliente.JogoID != "" {
		partidasMutex.Lock()
		partida, ok := partidas[cliente.JogoID]
		if ok {
			for _, jogador := range partida.Jogadores {
				if jogador.Cliente.Nome != cliente.Nome {
					codificadorAdversario := json.NewEncoder(jogador.Cliente.Conexao)
					mensagem := fmt.Sprintf("%s saiu do jogo, a partida acabou", cliente.Nome)
					servUtils.EnviarSauiPartida(*codificadorAdversario, mensagem)
					jogador.Cliente.Jogador = nil
					jogador.Cliente.JogoID = ""
					jogador.Cliente.Estado = ""
				}
			}
			delete(partidas, cliente.JogoID)
			cliente.JogoID = ""
			cliente.Estado = ""
		}
		partidasMutex.Unlock()

	}

	clientesMutex.Lock()
	delete(clientes, cliente.Nome)
	clientesMutex.Unlock()

	cliente.Conexao.Close()
	msg:=fmt.Sprintf("Cliente desconectado: %s\n", cliente.Nome)
	estilo.PrintVerm(msg)

}

func lidarComJogada(cliente *servUtils.Cliente, jogada protocolo.Jogada, codificador json.Encoder) {
	partidasMutex.Lock()
	partida, ok := partidas[cliente.JogoID]
	defer partidasMutex.Unlock()
	if ok {
		var adversario *servUtils.Jogador
		var codificadorAdiversario *json.Encoder
		for _, jogador := range partida.Jogadores {
			if jogador.Cliente.Nome != cliente.Nome {
				adversario = jogador
				codificadorAdiversario = json.NewEncoder(adversario.Cliente.Conexao)
			}
		}
		if partida.Turno == cliente.Nome {
			switch jogada.Acao {
			case "pegarCarta":
				//pode separar em uma função
				if len(partida.Cartas) > 0{
					carta := partida.Cartas[0]
					partida.Cartas = partida.Cartas[1:]
					cliente.Jogador.Mao = append(cliente.Jogador.Mao, carta)
					pontos := cartasUtils.TradutorPontos(carta)
					cliente.Jogador.Pontos += pontos
					servUtils.EnviarResJogada(codificador, carta, pontos, cliente)

					servUtils.EnviarAviso(*codificadorAdiversario, "Seu adversário pegou uma carta")
					if !adversario.ParouCartas {
						partida.Turno = adversario.Cliente.Nome
						servUtils.EnviarAviso(*codificadorAdiversario, " É o seu turno!")
					} else {
						servUtils.EnviarAviso(codificador, " É o seu turno!")
					}
				} else{
					servUtils.EnviarAviso(codificador, "Não há mais cartas")
					finalizarPartida(partida, cliente, adversario, &codificador, codificadorAdiversario)
				}
			case "pararCartas":
				cliente.Jogador.ParouCartas = true
				servUtils.EnviarAviso(codificador, "Você parou de pegar cartas")
				if adversario.ParouCartas {
					//pode separar em uma função
					//decide o vencedor, avisa quem foi e finaliza partida
					finalizarPartida(partida, cliente, adversario, &codificador, codificadorAdiversario)
				} else {
					servUtils.EnviarAviso(*codificadorAdiversario, " É o seu turno!")
					servUtils.EnviarAviso(*codificadorAdiversario, "Seu adversário parou de pegar cartas")
					partida.Turno = adversario.Cliente.Nome
				}
			}
		} else{
			servUtils.EnviarAviso(codificador, "❌ Não é o seu turno! ❌")
		}
	}
}

func finalizarPartida(partida *servUtils.Partida, cliente *servUtils.Cliente, adversario *servUtils.Jogador, codificador *json.Encoder, codificadorAdiversario *json.Encoder){
	pontosFinais := map[string]int{
		cliente.Nome:            cliente.Jogador.Pontos,
		adversario.Cliente.Nome: adversario.Pontos,
	}
	dis21cliente := int(math.Abs(float64(cliente.Jogador.Pontos) - 21))
	dis21adversario := int(math.Abs(float64(adversario.Pontos) - 21))
	if dis21cliente < dis21adversario {
		servUtils.EnviarFimPartida(codificador, codificadorAdiversario, cliente.Nome, pontosFinais)
	} else if dis21cliente > dis21adversario {
		servUtils.EnviarFimPartida(codificador, codificadorAdiversario, adversario.Cliente.Nome, pontosFinais)
	} else {
		servUtils.EnviarFimPartida(codificador, codificadorAdiversario, "empate", pontosFinais)
	}
	fecharPartida(partida)
}

func fecharPartida(partida *servUtils.Partida) {
	for _, jogador := range partida.Jogadores {
		estilo.PrintMag("FIM DE PARTIDA")
		jogador.Cliente.Jogador = nil
		jogador.Cliente.JogoID = ""
		jogador.Cliente.Estado = ""
	}
}