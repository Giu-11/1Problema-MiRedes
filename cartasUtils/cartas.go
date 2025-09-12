package cartasUtils

import (
	"math/rand"
	"sync"
)

// variaveis do estoque central de cartas
var totalCartas int
var estoqueCartas = make(map[string]map[string]int)
var cartasMutex = &sync.Mutex{}

// guada a pontuação de cada carta
var pontuacoes = map[string]int{
	"K":  10,
	"Q":  10,
	"J":  10,
	"10": 10,
	"9":  9,
	"8":  8,
	"7":  7,
	"6":  6,
	"5":  5,
	"4":  4,
	"3":  3,
	"2":  2,
	"A":  1}

// mostra quantos pontos cada carta vale
func TradutorPontos(carta string) int {
	return pontuacoes[carta]
}

// prepara um baralho embaralhado para uma partida
func GeradorCartasEmbaralhadas() []string {
	cartas := []string{"A", "K", "Q", "J", "10", "9", "8", "7", "6", "5", "4", "3", "2"}

	rand.Shuffle(len(cartas), func(i, j int) {
		cartas[i], cartas[j] = cartas[j], cartas[i]
	})

	return cartas
}

// inicia o estoque com uma quantidade fixa de cartas de cada valor e naipe
func CriadorEstoque() {
	cartasMutex.Lock()
	defer cartasMutex.Unlock()

	estoqueCartas = make(map[string]map[string]int)
	valores := []string{"A", "K", "Q", "J", "10", "9", "8", "7", "6", "5", "4", "3", "2"}
	//quantidade de cartas criada para cada naipe
	//raridade é baseada na propria quantidade de cartas
	quantidadesPorNaipe := map[string]int{
		"♥️": 10,
		"♠️": 20,
		"♦️": 30,
		"♣️": 40,
	}

	for _, valor := range valores {
		estoqueCartas[valor] = make(map[string]int)
		for naipe, quantidade := range quantidadesPorNaipe {
			estoqueCartas[valor][naipe] = quantidade
			totalCartas += quantidade
		}
	}
}

// sorteia uma carta para o cliente, retorna "" caso não hajam mais"
func AbrirPacote() (string, string) {
	cartasMutex.Lock()
	defer cartasMutex.Unlock()

	if totalCartas == 0 {
		return "", ""
	}

	r := rand.Intn(totalCartas)

	acumulado := 0
	for valor, naipes := range estoqueCartas {
		for naipe, qtd := range naipes {
			if qtd > 0 {
				acumulado += qtd
				if r < acumulado {
					estoqueCartas[valor][naipe]--
					totalCartas--
					return valor, naipe
				}
			}
		}
	}
	return "", ""
}
