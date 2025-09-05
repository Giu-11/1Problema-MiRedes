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

			time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

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

			conexao.SetReadDeadline(time.Now().Add(5 * time.Second))
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
			
			conexao.SetReadDeadline(time.Now().Add(10 * time.Second))
			
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