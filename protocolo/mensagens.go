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

type Mensagem struct{ //temporario, usado enquanto não ha funções de jogo
	Mensagem string
	Remetente string
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
	Resultado string
	pontos map[string]int
	//skins map[string](map[string]string)
	//ou
	//maos map[string]([]string)
}
