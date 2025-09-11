package cartasUtils

import (
	"math/rand"
	"sync"
)

type Carta struct {
	Naipe      string
	Tipo       string
	Quantidade int
}

var totalCartas int
var estoqueCartas = make(map[string]map[string]int)
var cartasMutex = &sync.Mutex{}

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

func TradutorPontos(carta string) int {
	return pontuacoes[carta]
}

func GeradorCartasEmbaralhadas() []string {
	cartas := []string{"A", "K", "Q", "J", "10", "9", "8", "7", "6", "5", "4", "3", "2"}

	rand.Shuffle(len(cartas), func(i, j int) {
		cartas[i], cartas[j] = cartas[j], cartas[i]
	})

	return cartas
}

func CriadorEstoque() {
	cartasMutex.Lock()
	defer cartasMutex.Unlock()

	estoqueCartas = make(map[string]map[string]int)
	valores := []string{"A", "K", "Q", "J", "10", "9", "8", "7", "6", "5", "4", "3", "2"}
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
