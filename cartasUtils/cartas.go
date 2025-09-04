package cartasUtils

import (
	"math/rand"
)

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
