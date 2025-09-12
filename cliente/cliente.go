package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"projeto-rede/estilo"
	"projeto-rede/protocolo"
	"sort"
	"strconv"
	"strings"
	"time"
)

//canais para conexão entre main input do usuario e servidor
var mensagensDoServidor = make(chan protocolo.Envelope)
var inputDoUsuario = make(chan string)

//cartas do cliente, para que não precise pedir sempre ao servidor
var inventarioCliente map[string]map[string]int

//deck de cartas do cliente
var deckEscolhido = make(map[string]string)

//variaveis para auxilio na montagem do deck
var valoresParaEscolher []string
var valorAtualParaEscolha string

func main() {
    serverAddress := "localhost:8080" 
    //se conecta a um endereço IP especifico caso o argumento seja passado
	if len(os.Args) > 1 {
        serverAddress = os.Args[1]
    }
    fmt.Printf("Tentando conectar ao servidor em %s...\n", serverAddress)


    conexao, err := net.Dial("tcp", serverAddress)
    if err != nil {
        fmt.Println("Erro ao conectar:", err)
        return
    }
    defer conexao.Close()

	codificador := json.NewEncoder(conexao)

	go receberMensagens(conexao)
	go lerInputDoUsuario()

	var nomeTemp string
	var nomeUsuario string
	estadoCliente := "login"
	estilo.Clear()
	fmt.Println("\n--- Bem-vindo! ---")
	fmt.Println("Digite seu nome de usuário para fazer o login:")
	fmt.Print(">> ")

	var startPing time.Time

	sair := false
	for !sair {
		select {
			//ouve as conexões para tratar com informações vindads delas
		case msgServidor := <-mensagensDoServidor:
			switch msgServidor.Requisicao {
				//trata mensagens do servidor
			case "confirmacao":
				var conf protocolo.Confirmacao
				json.Unmarshal(msgServidor.Dados, &conf)
				switch conf.Assunto {
				case "login": //caso o servidor mande uma resposta sobre a tentativa de login
					if conf.Resultado {
						estilo.Clear()
						nomeUsuario = nomeTemp
						estilo.PrintVerd("\n✅ Login realizado com sucesso!\n")
						estadoCliente = "menu"
						exibirMenu()
					} else {
						estilo.PrintVerm("\n⚠️ Falha no login. Tente outro nome.\n")
						estadoCliente = "login"
					}
				case "pacote": //caso o servidor mande um aviso de fim de cartas no estoque
					if !conf.Resultado {
						estilo.PrintVerm("Não há mais pacotes!❌")
					}
				}

			case "inicioPartida":
				estilo.Clear()
				var dadosPartida protocolo.InicioPartida
				json.Unmarshal(msgServidor.Dados, &dadosPartida)
				fmt.Printf("\n🃏--- PARTIDA INICIADA ---🃏\n")
				var primeiro string
				if nomeUsuario == dadosPartida.PrimeiroJogar{
					primeiro = "VOCÊ"
				} else{
					primeiro = dadosPartida.PrimeiroJogar
				}
				fmt.Printf("Contra: %s | Primeiro a jogar: %s\n", dadosPartida.Oponente, primeiro)
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
				naipe := deckEscolhido[resJogada.Carta]
				/*if naipe == "" { 
					naipe = "🚫"
				}*/
				fmt.Printf("Você conseguiu um %s%s\n+%d pontos!\nTotal de pontos:%d\n", resJogada.Carta, naipe, resJogada.PontosCarta, resJogada.PontosTotal)


			case "fimPartida":
				estilo.Clear()
				var dadosPartida protocolo.FimPartida
				json.Unmarshal(msgServidor.Dados, &dadosPartida)
				for nome, pontos := range dadosPartida.Pontos {
					fmt.Printf("%s conseguiu: ", nome)
					for _,cartas := range dadosPartida.Maos[nome]{
						if cartas != dadosPartida.Skins[nome][cartas]{
							fmt.Printf(" %s%s ",cartas, dadosPartida.Skins[nome][cartas])
						}else{
							fmt.Printf(" %s ",cartas)
						}
						
					}
					fmt.Print("\n")
					fmt.Printf("\t%s: %d pontos\n\n", nome, pontos)
				}
				if dadosPartida.Ganhador != "empate" {
					var ganhador string
					if nomeUsuario == dadosPartida.Ganhador{
						ganhador = "VOCÊ"
					}else{
						ganhador = dadosPartida.Ganhador
					}
					msg := fmt.Sprintf("\t\t%s GANHOU🎉!\n", ganhador)
					estilo.PrintVerd(msg)
				} else {
					fmt.Println("\t\tEMPATE")
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

			case "novaCarta":
				estilo.Clear()
				var dadosCarta protocolo.CartaNova
				json.Unmarshal(msgServidor.Dados, &dadosCarta)
				msg := fmt.Sprintf("Você obteve: %s%s", dadosCarta.Valor, dadosCarta.Naipe)
				estadoCliente = "menu"
				estilo.PrintCian(msg)
				exibirMenu()

			case "todasCartas":
				estilo.Clear()
				var cartas protocolo.TodasCartas
				json.Unmarshal(msgServidor.Dados, &cartas)
				inventarioCliente = cartas.Cartas 
				if estadoCliente == "preparandoDeck" {
					iniciarMontagemDeck()
					estadoCliente = "montandoDeck"
					fimEscolha := proximaEscolhaDeDeck() 
					if fimEscolha{
						estadoCliente = "menu"
						exibirMenu()
						fmt.Print(">> ")
					}

				} else {
					mostraCartas(inventarioCliente)
					exibirMenu()
					estadoCliente = "menu"
				}

			case "ping":
				if estadoCliente == "ping" {
					estilo.Clear()
					msg := fmt.Sprintf("🖥️Ping %s\n", time.Since(startPing))
					estilo.PrintAma(msg)
					estadoCliente = "menu"
					exibirMenu()
				}
			}
			if estadoCliente != "montandoDeck" {
				fmt.Print(">> ")
			}


		case input := <-inputDoUsuario:
			var msgParaEnviar protocolo.Envelope
			enviar := true

			switch estadoCliente {
			case "login":
				nomeTemp = strings.ToUpper(input)
				dados, _ := json.Marshal(protocolo.Login{Nome: nomeTemp})
				msgParaEnviar = protocolo.Envelope{Requisicao: "login", Dados: dados}
			case "menu":
				switch input {
				case "1": //entrar na fila de espera
					msgParaEnviar = protocolo.Envelope{Requisicao: "procurar"}
					estadoCliente = "esperando"
				case "2": //abrir pacote
					estadoCliente = "abrindoPacote"
					msgParaEnviar = protocolo.Envelope{Requisicao: "abrirPacote"}
				case "3": //ver ping
					estilo.Clear()
					msgParaEnviar = protocolo.Envelope{Requisicao: "ping"}
					startPing = time.Now()
					estadoCliente = "ping"
				case "4": //ver regras
					estilo.Clear()
					verRegras()
					exibirMenu()
					enviar = false
				case "5": // Ver todas cartas
					msgParaEnviar = protocolo.Envelope{Requisicao: "verCartas"}
					estadoCliente = "vendocartas"
				case "6": // Montar Deck
					msgParaEnviar = protocolo.Envelope{Requisicao: "verCartas"}
					estadoCliente = "preparandoDeck" // Estado intermediário
					fmt.Println("Buscando suas cartas para iniciar a montagem do deck...")
				case "7": //sair
					sair = true
					enviar = false
				default:
					estilo.Clear()
					estilo.PrintVerm("❌Opção inválida no menu.\n")
					exibirMenu()
					enviar = false
				}
			case "montandoDeck":
				enviar = false 
				naipesDisponiveis := getNaipesParaValor(valorAtualParaEscolha)

				escolha, err := strconv.Atoi(input)
				if err != nil || escolha < 1 || escolha > len(naipesDisponiveis) {
					estilo.PrintVerm("Escolha inválida, por favor digite um número da lista.\n")
				} else {
					naipeEscolhido := naipesDisponiveis[escolha-1]
					deckEscolhido[valorAtualParaEscolha] = naipeEscolhido
					estilo.PrintVerd(fmt.Sprintf("Você selecionou %s%s para a carta %s.\n", valorAtualParaEscolha, naipeEscolhido, valorAtualParaEscolha))
				}
				fimEscolha := proximaEscolhaDeDeck()
				if fimEscolha{
					enviar = true 
					dados,_ := json.Marshal(protocolo.NovoDeck{Deck: deckEscolhido})
					msgParaEnviar = protocolo.Envelope{Requisicao: "novoDeck", Dados: dados}
					estadoCliente = "menu"
					exibirMenu()
					fmt.Print(">> ")
				}

			case "jogando":
				switch input {
				case "1":
					dados, _ := json.Marshal(protocolo.Jogada{Acao: "pegarCarta"})
					msgParaEnviar = protocolo.Envelope{Requisicao: "jogada", Dados: dados}
				case "2":
					dados, _ := json.Marshal(protocolo.Jogada{Acao: "pararCartas"})
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
			} else if estadoCliente != "montandoDeck" {
				fmt.Print(">> ")
			}
		}
	}
	estilo.PrintCian("desconectando....\n até mais👋\n")
}

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

func mostraCartas(cartas map[string]map[string]int) {
	var valores []string
	for valor := range cartas {
		valores = append(valores, valor)
	}
	sort.Strings(valores)

	for _, valor := range valores {
		naipes := cartas[valor]
		fmt.Printf("\n")
		for naipe, quantidade := range naipes {
			fmt.Printf("%s%s x%d\t\t", valor, naipe, quantidade)
		}
	}
	fmt.Printf("\n")
}

func exibirMenu() {
	fmt.Println("\n--- MENU ---")
	fmt.Println("Digite:")
	fmt.Println("1-Procurar uma partida")
	fmt.Println("2-Abrir Pacote")
	fmt.Println("3-Ver PING")
	fmt.Println("4-Ver regras do jogo")
	fmt.Println("5-Ver suas cartas")
	fmt.Println("6-Montar seu Deck de Skins") 
	fmt.Println("7-Sair")                   
}

func exibirMenuPartida() {
	fmt.Println("\n\tSELECIONE SUA JOGADA!")
	fmt.Println("1-Pegar Carta")
	fmt.Println("2-Parar de pegar cartas")
}

func verRegras() {
	fmt.Println("\nEsse jogo é uma versão simplificada de 21")
	fmt.Println("Seu objetivo é conseguir o mais perto possivel de 21 pontos")
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


func iniciarMontagemDeck() {
	valoresParaEscolher = nil 
	for valor := range inventarioCliente {
		valoresParaEscolher = append(valoresParaEscolher, valor)
	}
	sort.Strings(valoresParaEscolher) 
	estilo.Clear()
	fmt.Println("--- MONTANDO SEU DECK ---")
	fmt.Println("Escolha um naipe (skin) para cada valor de carta que você possui.")
}

func proximaEscolhaDeDeck() bool{
	if len(valoresParaEscolher) > 0 {
		valorAtualParaEscolha = valoresParaEscolher[0]
		valoresParaEscolher = valoresParaEscolher[1:] 

		naipesDisponiveis := getNaipesParaValor(valorAtualParaEscolha)

		fmt.Printf("\nPara a carta de valor '%s', qual naipe você quer usar?\n", valorAtualParaEscolha)
		for i, naipe := range naipesDisponiveis {
			fmt.Printf("  %d: %s\n", i+1, naipe)
		}
		fmt.Print(">> ")
		return false
	} else {
		estilo.PrintVerd("\n🎉 Deck montado com sucesso! Suas escolhas foram salvas.\n")
		fmt.Println("Deck Atual:")
		for valor, naipe := range deckEscolhido {
			fmt.Printf("  %s -> %s%s\n", valor, valor, naipe)
		}
		return true
	}
}

func getNaipesParaValor(valor string) []string {
	var naipes []string
	if naipesDoValor, ok := inventarioCliente[valor]; ok {
		for naipe := range naipesDoValor {
			naipes = append(naipes, naipe)
		}
	}
	sort.Strings(naipes)
	return naipes
}

func envioDeckServidor(){
	//
}
