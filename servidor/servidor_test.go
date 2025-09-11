// Em servidor/servidor_test.go

package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"projeto-rede/protocolo"
	"sync"
	"testing"
	"time"
)

func TestMatchmakingConcorrente(t *testing.T) {
	const numClientes = 1200
	const addr = "localhost:8080"

	//Variavel responsavem por garantir que todas goroutines finalizem antes de mostrar os resultados
	var wg sync.WaitGroup
	wg.Add(numClientes)

	//map com a chave como nome de um jogador e valor o nome do oponente
	//para confirmar que mais de um jogador não entrou em duas partidas
	partidasEncontradas := make(map[string]string)
	//mutex para proteger o acesso ao placar
	var partidasMutex sync.Mutex

	for i := 0; i < numClientes; i++ {
		//Inicia cada cleinet simulado com uma goroutine diferente
		go func(id int) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			//conexão
			conexao, err := net.Dial("tcp", addr)
			if err != nil {
				t.Errorf("Cliente %d: falha ao conectar: %v", id, err)
				return
			}
			defer conexao.Close()

			codificador := json.NewEncoder(conexao)
			decodificador := json.NewDecoder(conexao)
			nome := fmt.Sprintf("Tester_%d", id)

			//login
			payloadLogin, _ := json.Marshal(protocolo.Login{Nome: nome})
			envelopeLogin := protocolo.Envelope{Requisicao: "login", Dados: payloadLogin}
			if err := codificador.Encode(envelopeLogin); err != nil {
				t.Errorf("Cliente %d: erro ao enviar login: %v", id, err)
				return
			}

			conexao.SetReadDeadline(time.Now().Add(120 * time.Second))
			var msgConfirmacao protocolo.Envelope
			if err := decodificador.Decode(&msgConfirmacao); err != nil {
				t.Errorf("Cliente %d: erro ao receber confirmação de login: %v", id, err)
				return
			}

			var conf protocolo.Confirmacao
			json.Unmarshal(msgConfirmacao.Dados, &conf)
			if !conf.Resultado {
				t.Errorf("Cliente %d: falha no login (servidor recusou)", id)
				return
			}

			//Busca por sala
			envelopeProcurar := protocolo.Envelope{Requisicao: "procurar"}
			codificador.Encode(envelopeProcurar)

			conexao.SetReadDeadline(time.Now().Add(120 * time.Second))

			for {
				var msgServidor protocolo.Envelope
				if err := decodificador.Decode(&msgServidor); err != nil {
					t.Errorf("Cliente %d: erro ao esperar pela partida: %v", id, err)
					return
				}

				//Caso entre em uma partida, sucesso!
				if msgServidor.Requisicao == "inicioPartida" {
					var dadosPartida protocolo.InicioPartida
					if err := json.Unmarshal(msgServidor.Dados, &dadosPartida); err != nil {
						t.Errorf("Cliente %d: erro ao decodificar payload da partida: %v", id, err)
						return
					}

					//Logica para verificar as partidas cridas
					partidasMutex.Lock()

					//verifica se um cliente está em uma partida além dessa
					if _, ok := partidasEncontradas[nome]; ok {
						t.Errorf("ERRO DE LÓGICA GRAVE: Cliente %s foi colocado em uma segunda partida!", nome)
					}

					//atualiza o placar
					partidasEncontradas[nome] = dadosPartida.Oponente
					//t.Logf("%s x %s", nome, dadosPartida.Oponente)

					partidasMutex.Unlock()

					break
				}
			}

		}(i)
	}

	wg.Wait()
	t.Run("AnaliseFinalDasPartidas", func(t *testing.T) {
		//Verifica se todas clinetes entraram em uma partida
		if len(partidasEncontradas) != numClientes {
			t.Errorf("Número incorreto de resultados. Esperado: %d, Obtido: %d. Alguns clientes não registraram sua partida.", numClientes, len(partidasEncontradas))
		}

		//verifica se as paridas são reciprocas
		for jogador, oponente := range partidasEncontradas {
			oponenteDoOponente, ok := partidasEncontradas[oponente]
			if !ok {
				t.Errorf("Inconsistência: Jogador %s foi pareado com %s, mas %s não registrou nenhuma partida.", jogador, oponente, oponente)
			} else if oponenteDoOponente != jogador {
				t.Errorf("Inconsistência de par: %s acha que jogou com %s, mas %s acha que jogou com %s.", jogador, oponente, oponente, oponenteDoOponente)
			} /*else if oponenteDoOponente == jogador{
				fmt.Printf("Partida criada corretamente! %s x %s\n", jogador, oponente)
			}*/
		}
	})
	fmt.Println("Teste de concorrência finalizado.\n\nRESULTADOS:")
}

func TestAbrirPacotesConcorrente(t *testing.T) {
	//numero de clientes deve ser um divisor de 1300
	// divisores de 1300: 1, 2, 4, 5, 10, 13, 20, 25, 26, 50, 52, 65, 100, 130, 260, 325, 650 e 1300
	const numClientes = 260
	const cartasPorCliente = 1300 / numClientes
	const addr = "localhost:8080"

	var wg sync.WaitGroup
	wg.Add(numClientes)

	var tipos = []string{"A", "K", "Q", "J", "10", "9", "8", "7", "6", "5", "4", "3", "2"}
	cartasRecebidas := make(map[string]map[string]int)
	for _, tipo := range tipos {
		cartasRecebidas[tipo] = make(map[string]int)
	}
	var cartasMutex sync.Mutex

	for i := 0; i < numClientes; i++ {
		// Inicia cada cliente com uma goroutine diferente
		go func(id int) {
			defer wg.Done()

			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			// Conexão
			conexao, err := net.Dial("tcp", addr)
			if err != nil {
				t.Errorf("Cliente %d: falha ao conectar: %v", id, err)
				return
			}
			defer conexao.Close()

			codificador := json.NewEncoder(conexao)
			decodificador := json.NewDecoder(conexao)
			nome := fmt.Sprintf("Tester_%d", id)

			// Login
			payloadLogin, _ := json.Marshal(protocolo.Login{Nome: nome})
			envelopeLogin := protocolo.Envelope{Requisicao: "login", Dados: payloadLogin}
			if err := codificador.Encode(envelopeLogin); err != nil {
				t.Errorf("Cliente %d: erro ao enviar login: %v", id, err)
				return
			}

			// Espera pela confirmação de login
			conexao.SetReadDeadline(time.Now().Add(120 * time.Second))
			var msgConfirmacao protocolo.Envelope
			if err := decodificador.Decode(&msgConfirmacao); err != nil {
				t.Errorf("Cliente %d: erro ao receber confirmação de login: %v", id, err)
				return
			}

			// logica para abrir pacotes
			numCartasObtidas := 0
			estoqueAcabou := false

			for numCartasObtidas < cartasPorCliente && !estoqueAcabou {
				envelopeAbrir := protocolo.Envelope{Requisicao: "abrirPacote"}
				if err := codificador.Encode(envelopeAbrir); err != nil {
					t.Errorf("Cliente %d: erro ao pedir carta: %v", id, err)
					return
				}

				conexao.SetReadDeadline(time.Now().Add(120 * time.Second))
				var msgServidor protocolo.Envelope
				if err := decodificador.Decode(&msgServidor); err != nil {
					t.Errorf("Cliente %d: erro ao esperar pela resposta: %v", id, err)
					return
				}

				switch msgServidor.Requisicao {
				case "novaCarta":
					var dadosCarta protocolo.CartaNova
					if err := json.Unmarshal(msgServidor.Dados, &dadosCarta); err == nil {
						//t.Logf("cliente %d: pegou uma carta", id)
						cartasMutex.Lock()

						if _, ok := cartasRecebidas[dadosCarta.Valor]; !ok {
							cartasRecebidas[dadosCarta.Valor] = make(map[string]int)
						}

						cartasRecebidas[dadosCarta.Valor][dadosCarta.Naipe]++
						numCartasObtidas++

						cartasMutex.Unlock()

					}

				//garante que se o estoque acabar, o codigo continue
				case "confirmacao":
					var conf protocolo.Confirmacao
					if err := json.Unmarshal(msgServidor.Dados, &conf); err == nil {
						if conf.Assunto == "pacote" && !conf.Resultado {
							estoqueAcabou = true
						}
					}
				}
			}

		}(i)
	}

	wg.Wait()

	wg.Wait()
	fmt.Println("Teste de concorrência finalizado.\n\n--- CONTAGEM DE CARTAS RECEBIDAS ---")

	t.Run("VerificacaoCompletaDasCartas", func(t *testing.T) {

		// Define o que esperamos que exista no baralho
		valoresEsperados := []string{"A", "K", "Q", "J", "10", "9", "8", "7", "6", "5", "4", "3", "2"}
		contagemEsperadaPorNaipe := map[string]int{
			"♥️": 10,
			"♠️": 20,
			"♦️": 30,
			"♣️": 40,
		}

		totalCartasContadas := 0

		for _, valor := range valoresEsperados {
			for naipe, contagemEsperada := range contagemEsperadaPorNaipe {

				contagemObtida := cartasRecebidas[valor][naipe]

				if contagemObtida != contagemEsperada {
					t.Errorf("Contagem incorreta para %s%s. Esperado: %d, Obtido: %d",
						valor, naipe, contagemEsperada, contagemObtida)
				}

				totalCartasContadas += contagemObtida
			}
		}

		fmt.Printf("\nTOTAL DE CARTAS DISTRIBUÍDAS (de acordo com os clientes): %d\n", totalCartasContadas)

		totalEsperadoNoEstoque := 1300
		if totalCartasContadas > totalEsperadoNoEstoque {
			t.Errorf("ERRO DE LÓGICA GRAVE: Mais cartas foram distribuídas (%d) do que existem no estoque (%d)!",
				totalCartasContadas, totalEsperadoNoEstoque)
		}

		totalPedidoPelosClientes := numClientes * cartasPorCliente
		if totalCartasContadas < totalPedidoPelosClientes && totalCartasContadas < totalEsperadoNoEstoque {
			t.Logf("AVISO: Quantidade de cartas distribuidas(%d) é menor que o estoque(%d)", totalCartasContadas, totalEsperadoNoEstoque)
		}
	})
}
