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
	//inicia o servidor criando o estoque e esperando conexões
	estilo.Clear()
	cartasUtils.CriadorEstoque()
	fmt.Println("Estoque de cartas gerado")
	fmt.Println("Servidor iniciado, aguardando conexões na porta 8080...")
	ouvinte, err := net.Listen("tcp", ":8080")
	if err != nil {
		msg := fmt.Sprintf("Erro ao iniciar o servidor:%s\n", err)
		estilo.PrintVerm(msg)
		return
	}
	defer ouvinte.Close()

	for {
		//inicia uma nova goroutine para os novos clientes conectados
		conexao, err := ouvinte.Accept()
		if err != nil {
			msg := fmt.Sprintf("Erro ao aceitar conexão:%s\n", err)
			estilo.PrintVerm(msg)
			continue
		}
		go lidarComConexao(conexao)
	}
}

func lidarComConexao(conexao net.Conn) {
	msg := fmt.Sprintf("Novo cliente conectado:%s\n", conexao.RemoteAddr().String())
	estilo.PrintVerd(msg)
	decodificador := json.NewDecoder(conexao)
	codificador := json.NewEncoder(conexao)

	//inicia um objeto cliente com skins e dados padão
	cliente := &servUtils.Cliente{Conexao: conexao, Nome: "não logado", Estado: "login", JogoID: ""}
	cliente.Skins = map[string]string{
		"K":  "K",
		"Q":  "Q",
		"J":  "J",
		"10": "10",
		"9":  "9",
		"8":  "8",
		"7":  "7",
		"6":  "6",
		"5":  "5",
		"4":  "4",
		"3":  "3",
		"2":  "2",
		"A":  "A",
	}

	defer desconectarCliente(cliente)

	sair := false
	for !sair {
		//le as mensagens do cliente
		var envelope protocolo.Envelope
		err := decodificador.Decode(&envelope)
		if err != nil {
			cliente.Mutex.Lock()
			nomeCliente := cliente.Nome
			cliente.Mutex.Unlock()
			if nomeCliente == "" {
				nomeCliente = "não logado"
			}
			fmt.Printf("Cliente %s perdeu a conexão. Erro: %v\n", nomeCliente, err)
			sair = true
			continue
		}

		//interpreta o pedido do cliente
		switch envelope.Requisicao {
		case "login":
			login(cliente, &envelope, codificador)

		case "procurar":
			cliente.Mutex.Lock()
			if cliente.Estado == "menu" {
				cliente.Estado = "esperando"
				cliente.Mutex.Unlock()
				addFilaEspera(cliente)
			} else {
				cliente.Mutex.Unlock()
				servUtils.EnviarAviso(*codificador, "Ação inválida")
			}

		case "jogada":
			cliente.Mutex.Lock()
			estado := cliente.Estado
			cliente.Mutex.Unlock()
			if estado == "jogando" {
				var jogada protocolo.Jogada
				if err := json.Unmarshal(envelope.Dados, &jogada); err == nil {
					lidarComJogada(cliente, jogada, *codificador)
				}
			}

		case "ping":
			if cliente.Estado == "menu" {
				resposta := protocolo.Envelope{Requisicao: "ping"}
				err := codificador.Encode(resposta)
				if err != nil {
					fmt.Println("Erro no envio de dados")
				}
			}

		case "abrirPacote":
			cliente.Mutex.Lock()
			estado := cliente.Estado
			cliente.Mutex.Unlock()

			if estado == "menu" {
				valor, naipe := cartasUtils.AbrirPacote()
				if valor != "" && naipe != "" {
					addCartaCliente(valor, naipe, cliente)
					servUtils.EnviarNovaCarta(codificador, valor, naipe)
				} else {
					servUtils.EnviarConfirmacao(*codificador, "confirmacao", "pacote", false)
				}
			}

		case "verCartas":
			//envia as cartas do estoque do cliente
			cliente.Mutex.Lock()
			cartas := cliente.Cartas
			cliente.Mutex.Unlock()
			servUtils.EnviarCartas(codificador, cartas)

		case "novoDeck":
			//atualiza o deck do cliente
			var novoDeck protocolo.NovoDeck
			if err := json.Unmarshal(envelope.Dados, &novoDeck); err == nil {
				mudarSkins(cliente, novoDeck.Deck)
			}
		}
	}
}

// lida com dados de recebimeto e envio de resultado do login
func login(cliente *servUtils.Cliente, envelope *protocolo.Envelope, codificador *json.Encoder) {
	cliente.Mutex.Lock()
	defer cliente.Mutex.Unlock()

	if cliente.Estado == "login" {
		var dadosLogin protocolo.Login
		if err := json.Unmarshal(envelope.Dados, &dadosLogin); err == nil {
			if tentarLogin(dadosLogin) {
				cliente.Nome = dadosLogin.Nome
				cliente.Estado = "menu"
				addListaClientes(cliente)
				servUtils.EnviarConfirmacao(*codificador, "confirmacao", "login", true)
			} else {
				servUtils.EnviarConfirmacao(*codificador, "confirmacao", "login", false)
			}
		}
	} else {
		servUtils.EnviarAviso(*codificador, "Ação inválida")
	}
}

// verifica se o nome é válido e se ele não existe na lista de clientes
func tentarLogin(dadosLogin protocolo.Login) bool {
	if dadosLogin.Nome != "" {
		clientesMutex.Lock()
		defer clientesMutex.Unlock()
		_, existe := clientes[dadosLogin.Nome]
		return !existe
	}
	return false
}

// adiciona um cliente a lista de clientes conecatados
func addListaClientes(cliente *servUtils.Cliente) {
	clientesMutex.Lock()
	defer clientesMutex.Unlock()
	clientes[cliente.Nome] = cliente
}

// adiciona um cliente a lista de espera por partida
func addFilaEspera(cliente *servUtils.Cliente) {
	esperaMutex.Lock()
	filaEspera = append(filaEspera, cliente)
	esperaMutex.Unlock()

	codificador := json.NewEncoder(cliente.Conexao)
	servUtils.EnviarAviso(*codificador, "Buscando adversário, aguarde...")
	fmt.Printf("Jogador %s entrou na fila de espera\n", cliente.Nome)
	verificarEspera()
}

// verifica se há clientes na fila de espera para criar uma paratida, se tiver, cria e avisa os clientes
func verificarEspera() {
	esperaMutex.Lock()
	if len(filaEspera) < 2 {
		esperaMutex.Unlock()
		return
	}

	cliente1 := filaEspera[0]
	cliente2 := filaEspera[1]
	filaEspera = filaEspera[2:]
	esperaMutex.Unlock()

	cliente1.Mutex.Lock()
	defer cliente1.Mutex.Unlock()
	cliente2.Mutex.Lock()
	defer cliente2.Mutex.Unlock()

	partidasMutex.Lock()
	defer partidasMutex.Unlock()

	jogoID := "jogo:" + strconv.FormatInt(time.Now().UnixNano(), 10)
	partida := &servUtils.Partida{ID: jogoID, Jogadores: make(map[string]*servUtils.Jogador)}

	jogador1 := &servUtils.Jogador{Cliente: cliente1}
	jogador2 := &servUtils.Jogador{Cliente: cliente2}

	cliente1.Jogador = jogador1
	cliente2.Jogador = jogador2
	cliente1.Estado = "jogando"
	cliente2.Estado = "jogando"
	cliente1.JogoID = jogoID
	cliente2.JogoID = jogoID

	partidas[jogoID] = partida
	partida.Jogadores[cliente1.Nome] = jogador1
	partida.Jogadores[cliente2.Nome] = jogador2
	partida.Cartas = cartasUtils.GeradorCartasEmbaralhadas()

	if rand.Intn(2) == 0 {
		partida.Turno = cliente1.Nome
	} else {
		partida.Turno = cliente2.Nome
	}

	msg := fmt.Sprintf("INICIANDO PARTIDA: %s x %s\n", cliente1.Nome, cliente2.Nome)
	estilo.PrintCian(msg)

	codificador1 := json.NewEncoder(cliente1.Conexao)
	codificador2 := json.NewEncoder(cliente2.Conexao)
	servUtils.EnviarInicioPartida(*codificador1, cliente2.Nome, partida.Turno)
	servUtils.EnviarInicioPartida(*codificador2, cliente1.Nome, partida.Turno)
}

// desconecta o cliente, removendo da lista de conexões e de possiveis partidas e fila de espera onde ele pode estar
func desconectarCliente(cliente *servUtils.Cliente) {
	cliente.Mutex.Lock()
	nomeCliente := cliente.Nome
	jogoID := cliente.JogoID
	cliente.Mutex.Unlock()

	// Remove da fila de espera
	esperaMutex.Lock()
	novaFila := []*servUtils.Cliente{}
	for _, c := range filaEspera {
		if c != cliente {
			novaFila = append(novaFila, c)
		}
	}
	filaEspera = novaFila
	esperaMutex.Unlock()

	// Se estava em jogo, notifica o oponente e limpa a partida
	if jogoID != "" {
		var oponente *servUtils.Cliente

		partidasMutex.Lock()
		partida, ok := partidas[jogoID]
		if ok {
			for _, jogador := range partida.Jogadores {
				if jogador.Cliente.Nome != nomeCliente {
					oponente = jogador.Cliente
				}
			}
			delete(partidas, jogoID)
		}
		partidasMutex.Unlock()

		if oponente != nil {
			oponente.Mutex.Lock()
			codificadorAdversario := json.NewEncoder(oponente.Conexao)
			mensagem := fmt.Sprintf("%s saiu do jogo, a partida acabou", nomeCliente)
			servUtils.EnviarSauiPartida(*codificadorAdversario, mensagem)
			oponente.Jogador = nil
			oponente.JogoID = ""
			oponente.Estado = "menu"
			oponente.Mutex.Unlock()
		}
	}

	// Remove da lista global de clientes
	if nomeCliente != "" && nomeCliente != "none" {
		clientesMutex.Lock()
		delete(clientes, nomeCliente)
		clientesMutex.Unlock()
	}

	cliente.Conexao.Close()
	msg := fmt.Sprintf("Cliente desconectado: %s\n", nomeCliente)
	estilo.PrintVerm(msg)
}

// interpreta e realiza a jogada do cliente, atualizando turno e avisando da jogada a todos
func lidarComJogada(cliente *servUtils.Cliente, jogada protocolo.Jogada, codificador json.Encoder) {
	partidasMutex.Lock()
	partida, ok := partidas[cliente.JogoID]
	partidasMutex.Unlock()

	if ok {
		partidasMutex.Lock()
		turnoAtual := partida.Turno
		var adversario *servUtils.Jogador
		for _, jogador := range partida.Jogadores {
			if jogador.Cliente.Nome != cliente.Nome {
				adversario = jogador
			}
		}
		partidasMutex.Unlock()

		if turnoAtual == cliente.Nome {
			switch jogada.Acao {
			case "pegarCarta":
				pegarCarta(partida, cliente, adversario, codificador, *json.NewEncoder(adversario.Cliente.Conexao))
			case "pararCartas":
				pararCartas(cliente, adversario, partida, &codificador, json.NewEncoder(adversario.Cliente.Conexao))
			}
		} else {
			servUtils.EnviarAviso(codificador, "❌ Não é o seu turno! ❌")
		}
	}
}

// realiza a jogada de pegar uma carta do baralho
func pegarCarta(partida *servUtils.Partida, cliente *servUtils.Cliente, adversario *servUtils.Jogador, codificador json.Encoder, codificadorAdiversario json.Encoder) {
	partidasMutex.Lock()
	defer partidasMutex.Unlock()

	if len(partida.Cartas) > 0 {
		carta := partida.Cartas[0]
		partida.Cartas = partida.Cartas[1:]
		cliente.Jogador.Mao = append(cliente.Jogador.Mao, carta)
		pontos := cartasUtils.TradutorPontos(carta)
		cliente.Jogador.Pontos += pontos

		if !adversario.ParouCartas {
			partida.Turno = adversario.Cliente.Nome
		}

		servUtils.EnviarResJogada(codificador, carta, pontos, cliente)
		servUtils.EnviarAviso(codificadorAdiversario, "Seu adversário pegou uma carta")
		if !adversario.ParouCartas {
			servUtils.EnviarAviso(codificadorAdiversario, " É o seu turno!")
		} else {
			servUtils.EnviarAviso(codificador, " É o seu turno!")
		}

	} else {
		servUtils.EnviarAviso(codificador, "Não há mais cartas")
		finalizarPartida(partida, cliente, adversario, &codificador, &codificadorAdiversario)
	}
}

// atualiza dados do cliente caso ele escolha não pegar mais cartas
func pararCartas(cliente *servUtils.Cliente, adversario *servUtils.Jogador, partida *servUtils.Partida, codificador *json.Encoder, codificadorAdiversario *json.Encoder) {

	cliente.Jogador.ParouCartas = true

	if adversario.ParouCartas {
		finalizarPartida(partida, cliente, adversario, codificador, codificadorAdiversario)
	} else {
		partida.Turno = adversario.Cliente.Nome
		servUtils.EnviarAviso(*codificador, "Você parou de pegar cartas")
		servUtils.EnviarAviso(*codificadorAdiversario, " É o seu turno!")
		servUtils.EnviarAviso(*codificadorAdiversario, "Seu adversário parou de pegar cartas")
	}
}

// finaliza a partida, decidindo um ganhador e informando os do resultado
func finalizarPartida(partida *servUtils.Partida, cliente *servUtils.Cliente, adversario *servUtils.Jogador, codificador *json.Encoder, codificadorAdiversario *json.Encoder) {
	pontosFinais := map[string]int{
		cliente.Nome:            cliente.Jogador.Pontos,
		adversario.Cliente.Nome: adversario.Pontos,
	}
	vencedor := "empate"
	dis21cliente := int(math.Abs(float64(cliente.Jogador.Pontos) - 21))
	dis21adversario := int(math.Abs(float64(adversario.Pontos) - 21))
	if dis21cliente < dis21adversario {
		vencedor = cliente.Nome
	} else if dis21cliente > dis21adversario {
		vencedor = adversario.Cliente.Nome
	}
	maos := map[string][]string{
		cliente.Nome:            cliente.Jogador.Mao,
		adversario.Cliente.Nome: adversario.Mao,
	}

	skins := map[string]map[string]string{
		cliente.Nome:            cliente.Skins,
		adversario.Cliente.Nome: adversario.Cliente.Skins,
	}

	servUtils.EnviarFimPartida(codificador, codificadorAdiversario, vencedor, pontosFinais, maos, skins)
	fecharPartida(partida)
}

// tira a partida do
func fecharPartida(partida *servUtils.Partida) {
	for _, jogador := range partida.Jogadores {
		jogador.Cliente.Mutex.Lock()
		jogador.Cliente.Jogador = nil
		jogador.Cliente.JogoID = ""
		jogador.Cliente.Estado = "menu"
		jogador.Cliente.Mutex.Unlock()
	}
	partidasMutex.Lock()
	delete(partidas, partida.ID)
	estilo.PrintMag("FIM DE PARTIDA\n")
	partidasMutex.Unlock()
}

// adiciona uma nova carta ao estoque do cliente
func addCartaCliente(valor string, naipe string, cliente *servUtils.Cliente) {
	cliente.Mutex.Lock()
	defer cliente.Mutex.Unlock()

	if cliente.Cartas == nil {
		cliente.Cartas = make(map[string]map[string]int)
	}

	if cliente.Cartas[valor] == nil {
		cliente.Cartas[valor] = make(map[string]int)
	}

	cliente.Cartas[valor][naipe]++
}

// atualiza o deck de skins do cliente
func mudarSkins(cliente *servUtils.Cliente, skins map[string]string) {
	cliente.Mutex.Lock()
	defer cliente.Mutex.Unlock()
	for carta := range cliente.Skins {
		cliente.Skins[carta] = skins[carta]
	}
}
