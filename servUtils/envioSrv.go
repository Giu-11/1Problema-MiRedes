package servUtils

import (
	"encoding/json"
	"fmt"
	"net"
	"projeto-rede/protocolo"
	"sync"
)

// cliente dentro do servidor
type Cliente struct {
	Conexao net.Conn //conexão do cliente
	Nome    string
	JogoID  string
	Estado  string
	Jogador *Jogador                  //objeto jogador quando o cliente estiver dentro de uma partida
	Skins   map[string]string         //deck de skins
	Cartas  map[string]map[string]int //estoque de cartas do cliente
	Mutex   sync.Mutex                //mutex para mudanças do cliente
}

// cliente dentro de uma partida
type Jogador struct {
	Cliente     *Cliente // cliente que esse jogador se refere
	Mao         []string //cartas que ele pegou do baralho
	Pontos      int
	ParouCartas bool //se ele já parou de pegar cartas
}

// partida entre dosi clientes
type Partida struct {
	ID        string              //identificação dela
	Jogadores map[string]*Jogador //referencia aos jogadores da partida
	Turno     string              //quem deve jogar
	Cartas    []string            //baralho de cartas
}

// envia uma resposta a jogada
func EnviarResJogada(codificador json.Encoder, carta string, pontos int, cliente *Cliente) {
	resposta := protocolo.Envelope{Requisicao: "resJogada"}
	respostaJogada := protocolo.RespostaJogada{Carta: carta, PontosCarta: pontos, PontosTotal: cliente.Jogador.Pontos}

	dadosCod, err := json.Marshal(respostaJogada)
	if err == nil {
		resposta.Dados = dadosCod
		err := codificador.Encode(resposta)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia uma confirmação a uma ação do cliente,
func EnviarConfirmacao(codificador json.Encoder, requisicao string, assunto string, resultado bool) {
	resposta := protocolo.Envelope{Requisicao: requisicao}
	respostaLogin := protocolo.Confirmacao{Assunto: assunto, Resultado: resultado}

	dadosCod, err := json.Marshal(respostaLogin)
	if err == nil {
		resposta.Dados = dadosCod
		err := codificador.Encode(resposta)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia um aviso
func EnviarAviso(codificador json.Encoder, aviso string) {
	resposta := protocolo.Envelope{Requisicao: "notfServidor"}
	respostaLogin := protocolo.Mensagem{Mensagem: aviso}

	dadosCod, err := json.Marshal(respostaLogin)
	if err == nil {
		resposta.Dados = dadosCod
		err := codificador.Encode(resposta)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia aviso que adiversario saiu da partida
func EnviarSauiPartida(codificador json.Encoder, mensagem string) {
	envelope := protocolo.Envelope{Requisicao: "saiuPartida"}
	respostaLogin := protocolo.Mensagem{Mensagem: mensagem}
	dadosCod, err := json.Marshal(respostaLogin)

	if err == nil {
		envelope.Dados = dadosCod
		err := codificador.Encode(envelope)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia dados sobre o fim da partida
func EnviarFimPartida(codificador *json.Encoder, codificador2 *json.Encoder, resultado string, pontos map[string]int, maos map[string][]string, skins map[string]map[string]string) {
	envelope := protocolo.Envelope{Requisicao: "fimPartida"}
	fimPartida := protocolo.FimPartida{Pontos: pontos, Ganhador: resultado, Skins: skins, Maos: maos}
	dadosCod, err := json.Marshal(fimPartida)

	if err == nil {
		envelope.Dados = dadosCod
		err := codificador.Encode(envelope)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
		err = codificador2.Encode(envelope)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia dados sobre o inicio de uma nova partida
func EnviarInicioPartida(codificador json.Encoder, oponente string, primeiroJogar string) {
	resposta := protocolo.Envelope{Requisicao: "inicioPartida"}
	respostaLogin := protocolo.InicioPartida{Oponente: oponente, PrimeiroJogar: primeiroJogar}

	dadosCod, err := json.Marshal(respostaLogin)
	if err == nil {
		resposta.Dados = dadosCod
		err := codificador.Encode(resposta)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia dados da carta q	ue o cliente obteve
func EnviarNovaCarta(codificador *json.Encoder, valor string, naipe string) {
	resposta := protocolo.Envelope{Requisicao: "novaCarta"}
	dadosCarta := protocolo.CartaNova{Valor: valor, Naipe: naipe}

	dadosCod, err := json.Marshal(dadosCarta)
	if err == nil {
		resposta.Dados = dadosCod
		err := codificador.Encode(resposta)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}

// envia estoque de cartas do cliente
func EnviarCartas(codificador *json.Encoder, cartas map[string]map[string]int) {
	resposta := protocolo.Envelope{Requisicao: "todasCartas"}
	dadosCarta := protocolo.TodasCartas{Cartas: cartas}

	dadosCod, err := json.Marshal(dadosCarta)
	if err == nil {
		resposta.Dados = dadosCod
		err := codificador.Encode(resposta)
		if err != nil {
			fmt.Println("Erro no envio de dados")
		}
	} else {
		fmt.Println("Erro de codificação de dados")
	}
}
