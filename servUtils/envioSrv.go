package servUtils

import (
	"encoding/json"
	"fmt"
	"net"
	"projeto-rede/protocolo"
	"sync"
)

type Cliente struct {
	Conexao net.Conn
	Nome    string
	JogoID  string
	Estado  string
	Jogador *Jogador
	Cartas  map[string]map[string]int
	Mutex   sync.Mutex
}

type Jogador struct {
	Cliente     *Cliente
	Mao         []string
	Pontos      int
	ParouCartas bool
	//skins map[string]*string //TODO: implemantar as esteticas de cartas
}

type Partida struct {
	ID        string
	Jogadores map[string]*Jogador
	Turno     string
	Cartas    []string
}

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

func EnviarResposta(codificador json.Encoder, requisicao string, assunto string, resultado bool) {
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

func EnviarFimPartida(codificador *json.Encoder, codificador2 *json.Encoder, resultado string, pontos map[string]int) {
	envelope := protocolo.Envelope{Requisicao: "fimPartida"}
	fimPartida := protocolo.FimPartida{Pontos: pontos, Ganhador: resultado}
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
