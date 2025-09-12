package protocolo

import "encoding/json"

type Envelope struct {
	Requisicao string
	Dados      json.RawMessage
}

type Login struct{
	Nome string
	Senha string
}

type InicioPartida struct{
	Oponente string
	PrimeiroJogar string
}

type Confirmacao struct{
	Assunto string
	Resultado bool
}

type Mensagem struct{ 
	Mensagem string
}

type Jogada struct{
	Acao string
}

type RespostaJogada struct{
	Carta string
	PontosCarta int
	PontosTotal int
}

type FimPartida struct{
	Ganhador string
	Pontos map[string]int
	Skins map[string](map[string]string)
	Maos map[string]([]string)
}

type CartaNova struct{
	Valor string
	Naipe string
}

type TodasCartas struct{
	Cartas map[string]map[string]int
}

type NovoDeck struct{
	Deck map[string]string
}
