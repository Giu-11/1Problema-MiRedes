package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
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
	fmt.Println("--- Bem-vindo! ---")
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
					fmt.Println("\nLogin realizado com sucesso!")
					estadoCliente = "menu"
					exibirMenu(estadoCliente)
				} else {
					fmt.Println("\nFalha no login. Tente outro nome.")
					estadoCliente = "login"
				}

			case "inicioPartida":
				var dadosPartida protocolo.InicioPartida
				json.Unmarshal(msgServidor.Dados, &dadosPartida)
				fmt.Printf("\n--- PARTIDA INICIADA ---\n")
				fmt.Printf("Contra: %s | Primeiro a jogar: %s\n", dadosPartida.Oponente, dadosPartida.PrimeiroJogar)
				fmt.Println("------------------------")
				estadoCliente = "jogando"

			case "notfServidor":
				var notif protocolo.Mensagem
				json.Unmarshal(msgServidor.Dados, &notif)
				fmt.Printf("\n--- SERVIDOR: %s ---\n", notif.Mensagem)

			case "mensagem":
				var msg protocolo.Mensagem
				json.Unmarshal(msgServidor.Dados, &msg)
				fmt.Printf("\n[%s]: %s\n", msg.Remetente, msg.Mensagem)
			
			case "saiuPartida":
				var msg protocolo.Mensagem
				json.Unmarshal(msgServidor.Dados, &msg) 
				fmt.Printf("\n--- %s ---\n", msg.Mensagem)
				estadoCliente = "menu"
				exibirMenu(estadoCliente)
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
					fmt.Println("Opção inválida no menu.")
					enviar = false
				}
			case "jogando":
				dados, _ := json.Marshal(protocolo.Mensagem{Mensagem: input})
				msgParaEnviar = protocolo.Envelope{Requisicao: "enviarmsg", Dados: dados}
			case "esperando":
				fmt.Println("Aguardando um adversário, por favor espere...")
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
			fmt.Println("\nConexão perdida com o servidor.")
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

func exibirMenu(estado string) {
	if estado == "menu" {
		fmt.Println("\n--- VOCÊ ESTÁ NO MENU ---")
		fmt.Println("Digite 'PROCURAR' para encontrar uma partida.")
	}
}