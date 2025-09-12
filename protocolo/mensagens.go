package protocolo

import "encoding/json"

//envelope geral contendo motivo da mensagem e dados adicionais
type Envelope struct {
	Requisicao string
	Dados      json.RawMessage
}

//dados do login
type Login struct {
	Nome string
	//Senha string
}

//dados iniciais da partida
type InicioPartida struct {
	Oponente      string
	PrimeiroJogar string
}

//manda uma confirmação de uma ação teve sucesso
type Confirmacao struct {
	Assunto   string
	Resultado bool
}

//manda uma mensagem de texto simples
type Mensagem struct {
	Mensagem string
}

//manda a jogada do jogador
type Jogada struct {
	Acao string
}

//resultados da jogada
type RespostaJogada struct {
	Carta       string
	PontosCarta int
	PontosTotal int
}

//informações do final da partida
type FimPartida struct {
	Ganhador string
	Pontos   map[string]int
	Skins    map[string](map[string]string)
	Maos     map[string]([]string)
}

//dados da nova carta do cliente
type CartaNova struct {
	Valor string
	Naipe string
}

//manda o estoque de cartas do cliente
type TodasCartas struct {
	Cartas map[string]map[string]int
}

//manda o deck escolhido pelo cliente
type NovoDeck struct {
	Deck map[string]string
}
