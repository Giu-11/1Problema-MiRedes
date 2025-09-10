package cartasUtils

import (
	"math/rand"
	"sync"
)

type Carta struct{
	Naipe string
	Tipo string
	Quantidade int
}

var tipos = []string{"A", "K", "Q", "J","10", "9", "8", "7","6", "5", "4", "3", "2",}
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

func CriadorEstoque(){
	cartasMutex.Lock()
	defer cartasMutex.Unlock()
	naipes := []string{"♥️", "♠️", "♦️", "♣️"}
	quantidades := []int{10, 20, 30, 40}
	for carta := range pontuacoes{
		estoqueCartas[carta] = make(map[string]int)
		for i := 0; i < 4; i++{
			estoqueCartas[carta][naipes[i]] = quantidades[i]
		}
	}
}

func AbrirPacote()(string, string){
	cartasMutex.Lock()
	defer cartasMutex.Unlock()
	 total := 0
    for _, naipes := range estoqueCartas {
        for _, qtd := range naipes {
            total += qtd
        }
    }
    if total == 0 {
        return "", ""
    }

    r := rand.Intn(total)

    acumulado := 0
    for valor, naipes := range estoqueCartas {
        for naipe, qtd := range naipes {
            acumulado += qtd
            if r < acumulado {
                estoqueCartas[valor][naipe]-- 
                return valor, naipe
            }
        }
    }
    return "", ""
}