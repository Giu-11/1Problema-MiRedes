package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"projeto-rede/estilo"
	"projeto-rede/protocolo"
	"strings"
)

var mensagensDoServidor = make(chan protocolo.Envelope)
var inputDoUsuario = make(chan string)

func main() {
	conexao, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar:", err)
		return
	}
	defer conexao.Close()

	codificador := json.NewEncoder(conexao)

	go receberMensagens(conexao)
	go lerInputDoUsuario()

	estadoCliente := "login"
	estilo.Clear()
	fmt.Println("\n--- Bem-vindo! ---")
	fmt.Println("Digite seu nome de usuário para fazer o login:")
	fmt.Print(">> ")

	for {
		select {
		case msgServidor := <-mensagensDoServidor:
			switch msgServidor.Requisicao {
			case "confirmacao":
				var conf protocolo.Confirmacao
				json.Unmarshal(msgServidor.Dados, &conf)
				if conf.Assunto == "login" && conf.Resultado {
					estilo.Clear()
					estilo.PrintVerd("\n✅ Login realizado com sucesso!\n")
					estadoCliente = "menu"
					exibirMenu()
				} else {
					estilo.PrintVerm("\n⚠️ Falha no login. Tente outro nome.\n")
					estadoCliente = "login"
				}

			case "inicioPartida":
				estilo.Clear()
				var dadosPartida protocolo.InicioPartida
				json.Unmarshal(msgServidor.Dados, &dadosPartida)
				fmt.Printf("\n🃏--- PARTIDA INICIADA ---🃏\n")
				fmt.Printf("Contra: %s | Primeiro a jogar: %s\n", dadosPartida.Oponente, dadosPartida.PrimeiroJogar)
				fmt.Println("------------------------")
				estadoCliente = "jogando"
				exibirMenuPartida()

			case "notfServidor":
				var notif protocolo.Mensagem
				json.Unmarshal(msgServidor.Dados, &notif)
				msg := fmt.Sprintf("\n--- %s ---\n", notif.Mensagem)
				estilo.PrintAma(msg)

			case "resJogada":
				var resJogada protocolo.RespostaJogada
				json.Unmarshal(msgServidor.Dados, &resJogada)
				fmt.Printf("Você conseguiu um %s\n+%d pontos!\nTotal de pontos:%d\n", resJogada.Carta, resJogada.PontosCarta, resJogada.PontosTotal)

			case "fimPartida":
				estilo.Clear()
				var dadosPartida protocolo.FimPartida
				json.Unmarshal(msgServidor.Dados, &dadosPartida)
				for nome, pontos := range dadosPartida.Pontos{
					fmt.Printf("%s conseguiu %d pontos\n", nome, pontos)
				}
				if dadosPartida.Ganhador != "empate"{
					msg := fmt.Sprintf("%s GANHOU🎉!\n", dadosPartida.Ganhador)
					estilo.PrintVerd(msg)
				} else{
					fmt.Println("EMPATE")
				}
				estadoCliente = "menu"
				exibirMenu()

			case "saiuPartida":
				estilo.Clear()
				var mensagem protocolo.Mensagem
				json.Unmarshal(msgServidor.Dados, &mensagem) 
				msg := fmt.Sprintf("\n--- ⚠️%s⚠️ ---\n", mensagem.Mensagem)
				estilo.PrintVerm(msg)
				estadoCliente = "menu"
				exibirMenu()
			}
			fmt.Print(">> ")

		case input := <-inputDoUsuario:
			var msgParaEnviar protocolo.Envelope
			enviar := true

			switch estadoCliente {
			case "login":
				dados, _ := json.Marshal(protocolo.Login{Nome: input})
				msgParaEnviar = protocolo.Envelope{Requisicao: "login", Dados: dados}
			case "menu":
				if strings.ToUpper(input) == "PROCURAR" {
					msgParaEnviar = protocolo.Envelope{Requisicao: "procurar"}
					estadoCliente = "esperando"
				} else {
					estilo.PrintVerm("❌Opção inválida no menu.\n")
					enviar = false
				}
			case "jogando":
				switch input {
				case "1":
					dados,_:= json.Marshal(protocolo.Jogada{Acao: "pegarCarta"})
					msgParaEnviar = protocolo.Envelope{Requisicao: "jogada", Dados: dados}
				case "2":
					dados,_:= json.Marshal(protocolo.Jogada{Acao: "pararCartas"})
					msgParaEnviar = protocolo.Envelope{Requisicao: "jogada", Dados: dados}
				default:
					estilo.PrintVerm("❌Opção inválida.\n")
				}
			case "esperando":
				fmt.Println("⌛Aguardando um adversário, por favor espere...⌛")
				enviar = false
			}

			if enviar {
				codificador.Encode(msgParaEnviar)
			} else {
				fmt.Print(">> ")
			}
		}
	}
}

// As goroutines de 'sentidos' continuam as mesmas
func receberMensagens(conexao net.Conn) {
	decodificador := json.NewDecoder(conexao)
	for {
		var msg protocolo.Envelope
		if err := decodificador.Decode(&msg); err != nil {
			estilo.PrintVerm("\n⛓️‍💥Conexão perdida com o servidor.")
			os.Exit(0)
		}
		mensagensDoServidor <- msg
	}
}

func lerInputDoUsuario() {
	leitor := bufio.NewReader(os.Stdin)
	for {
		input, _ := leitor.ReadString('\n')
		inputDoUsuario <- strings.TrimSpace(input)
	}
}

func exibirMenu() {
	fmt.Println("\n--- VOCÊ ESTÁ NO MENU ---")
	fmt.Println("Digite 'PROCURAR' para encontrar uma partida.")

}

func exibirMenuPartida(){
	fmt.Println("\n\tSELECIONE SUA JOGADA!")
	fmt.Println("1-Pegar Carta")
	fmt.Println("2-Parar de pegar cartas")
}

func verRegras(){
	fmt.Println("\nEsse jogo é uma versão simplificada de 21")
	fmt.Println("Seu objetivo é conseguir o mais perto possivil de 21 pontos")
	fmt.Println("As cartas valem:")
	fmt.Println("K: 10")
	fmt.Println("Q: 10")
	fmt.Println("J: 10")
	fmt.Println("10: 10")
	fmt.Println("9: 9")
	fmt.Println("8: 8")
	fmt.Println("7: 7")
	fmt.Println("6: 6")
	fmt.Println("5: 5")
	fmt.Println("4: 4")
	fmt.Println("3: 3")
	fmt.Println("2: 2")
	fmt.Println("A: 1")
	fmt.Println("Em cada turno você pode escolher pegar uma carta, ou parar de pegar cartas finalizando suas jogadas")
}