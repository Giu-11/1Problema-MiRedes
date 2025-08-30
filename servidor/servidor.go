package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"projeto-rede/protocolo"
	"strconv"
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
	decodificador := json.NewDecoder(conexao)
	codificador := json.NewEncoder(conexao)

	cliente := &Cliente{conexao: conexao, nome: "", estado: "login", jogoID: ""}

	defer desconectarCliente(cliente)

	sair := false
	for !sair{
		//
		var envelope protocolo.Envelope
		err := decodificador.Decode(&envelope)
		if err == nil{
			switch envelope.Requisicao{
			case"login":
				if cliente.estado=="login"{
					var dadosLogin protocolo.Login
					err:=json.Unmarshal(envelope.Dados, &dadosLogin)
					if err == nil{
						if tentarLogin(dadosLogin){
							cliente.nome=dadosLogin.Nome
							cliente.estado=""
							addListaClientes(cliente)
							enviarResposta(*codificador, "confirmacao", "login", true)
						}else{
							enviarResposta(*codificador, "confirmacao", "login", false)
						}
					}
				}else{
					enviarAviso(*codificador, "Ação inválida")
				}
			case "procurar":
				if cliente.estado == ""{
					cliente.estado="esperando"
					addFilaEspera(cliente)
					//TODO: função de sair da fila de espera
				}else{
					enviarAviso(*codificador, "Ação inválida")
				}
			case "enviarmsg":
				if cliente.estado=="jogando"{
					var mensagem protocolo.Mensagem
					err:=json.Unmarshal(envelope.Dados, &mensagem)
					if err == nil{
						lidarMensagem(cliente, mensagem.Mensagem, *codificador)
					}
				}else{
					enviarAviso(*codificador, "Ação inválida")
					fmt.Println(cliente.estado)
				}
			}
		}else{
			fmt.Printf("Cliente %s perdeu a conexão. Erro: %v\n", cliente.nome, err)
			sair = true
		}
	}
}

func tentarLogin(dadosLogin protocolo.Login)bool{
	//TODO: logica para login
	if dadosLogin.Nome !=""{
		return true
	}else{
		return false
	}
}

func addListaClientes(cliente *Cliente){
	clientesMutex.Lock()
	clientes[cliente.nome] = cliente
	clientesMutex.Unlock()
}

func addFilaEspera(cliente *Cliente) {
	cliente.estado = "esperando"
	esperaMutex.Lock()
	filaEspera = append(filaEspera, cliente)
	esperaMutex.Unlock()
	codificador := json.NewEncoder(cliente.conexao)
    enviarAviso(*codificador, "Buscando adversário, aguarde...")
	fmt.Printf("Jogador %s entrou na fila de espera\n", cliente.nome)
	verificarEspera()
}

func verificarEspera() {
	esperaMutex.Lock()
	defer esperaMutex.Unlock()

	if len(filaEspera) >= 2 {
		Cliente1 := filaEspera[0]
		Cliente2 := filaEspera[1]

		codificador1 := json.NewEncoder(Cliente1.conexao)
    	codificador2 := json.NewEncoder(Cliente2.conexao)

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
		clientesMutex.Lock()
		partidas[jogoID] = partida
		partida.jogadores[Cliente1.nome] = jogador1
		partida.jogadores[Cliente2.nome] = jogador2
		Cliente1.estado = "jogando"
		Cliente2.estado = "jogando"
		Cliente1.jogoID = jogoID
		Cliente2.jogoID = jogoID
		clientesMutex.Unlock()
		partidasMutex.Unlock()

		enviarInicioPartida(*codificador1, Cliente2.nome, partida.turno)
		enviarInicioPartida(*codificador2, Cliente1.nome, partida.turno)

		switch primeiroJogador{
		case 0:
			enviarAviso(*codificador1, "Você é o primeiro a jogar!")
        	enviarAviso(*codificador2, "Você é o segundo a jogar!")
		case 1:
			enviarAviso(*codificador2, "Você é o primeiro a jogar!")
        	enviarAviso(*codificador1, "Você é o segundo a jogar!")
		}
	}
}

func enviarInicioPartida(codificador json.Encoder, oponente string, primeiroJogar string){
	resposta := protocolo.Envelope{Requisicao: "inicioPartida"}
	respostaLogin := protocolo.InicioPartida{Oponente: oponente, PrimeiroJogar: primeiroJogar}

	dadosCod, err := json.Marshal(respostaLogin)
	if err==nil{
		resposta.Dados = dadosCod
		err:=codificador.Encode(resposta)
		if err != nil{
			fmt.Println("Erro no envio de dados")
		}
	}else{
		fmt.Println("Erro de codificação de dados")
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
					codificadorAdversario := json.NewEncoder(jogador.cliente.conexao)
					mensagem := fmt.Sprintf("%s saiu do jogo, a partida acabou", cliente.nome)
					enviarFimPartida(*codificadorAdversario, mensagem)
					//enviarAviso(*codificadorAdversario, mensagem)
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
	fmt.Printf("\033[31mCliente desconectado: %s\n\033[0m", cliente.nome)

}

func lidarMensagem(remetente *Cliente, mensagem string, codificador json.Encoder) {
	partidasMutex.Lock()
	partida, ok := partidas[remetente.jogoID]
	partidasMutex.Unlock()

	if ok {
		if partida.turno == remetente.nome{
			for _, destinatario := range partida.jogadores {
				if destinatario.cliente.nome != remetente.nome {
					codificadorDest := json.NewEncoder(destinatario.cliente.conexao)
					enviarMensagem(*codificadorDest, mensagem, remetente.nome)
					partida.turno = destinatario.cliente.nome
					enviarAviso(*codificadorDest, ">>> É o seu turno!")
					enviarAviso(codificador, ">>> Turno do seu oponente")
				}
			}

		}else{
			enviarAviso(codificador, "Não é seu turno")
		}
	}
}

func enviarMensagem(codificador json.Encoder, mensagem string, remetente string){
	resposta := protocolo.Envelope{Requisicao: "mensagem"}
	respostaLogin := protocolo.Mensagem{Remetente: remetente, Mensagem: mensagem}

	dadosCod, err := json.Marshal(respostaLogin)
	if err==nil{
		resposta.Dados = dadosCod
		err:=codificador.Encode(resposta)
		if err != nil{
			fmt.Println("Erro no envio de dados")
		}
	}else{
		fmt.Println("Erro de codificação de dados")
	}
}

func enviarResposta(codificador json.Encoder, requisicao string, assunto string, resultado bool){
	resposta := protocolo.Envelope{Requisicao: requisicao}
	respostaLogin := protocolo.Confirmacao{Assunto: assunto, Resultado: resultado}

	dadosCod, err := json.Marshal(respostaLogin)
	if err==nil{
		resposta.Dados = dadosCod
		err:=codificador.Encode(resposta)
		if err != nil{
			fmt.Println("Erro no envio de dados")
		}
	}else{
		fmt.Println("Erro de codificação de dados")
	}
}

func enviarAviso(codificador json.Encoder, aviso string){
	resposta := protocolo.Envelope{Requisicao: "notfServidor"}
	respostaLogin := protocolo.Mensagem{Mensagem: aviso}

	dadosCod, err := json.Marshal(respostaLogin)
	if err==nil{
		resposta.Dados = dadosCod
		err:=codificador.Encode(resposta)
		if err != nil{
			fmt.Println("Erro no envio de dados")
		}
	}else{
		fmt.Println("Erro de codificação de dados")
	}
}

func enviarFimPartida(codificador json.Encoder, mensagem string){
	envelope := protocolo.Envelope{Requisicao: "saiuPartida"}
	respostaLogin := protocolo.Mensagem{Mensagem: mensagem}
	dadosCod, err := json.Marshal(respostaLogin)

	if err==nil{
		envelope.Dados = dadosCod
		err:=codificador.Encode(envelope)
		if err != nil{
			fmt.Println("Erro no envio de dados")
		}
	}else{
		fmt.Println("Erro de codificação de dados")
	}
}
