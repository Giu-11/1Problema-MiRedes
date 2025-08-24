package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {

	conexao, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Erro ao conectar ao servidor:", err)
		return
	}
	defer conexao.Close()

	go receberMensagens(conexao)

	leitor := bufio.NewReader(os.Stdin)
	for {
		mensagem, _ := leitor.ReadString('\n')
		_, err := conexao.Write([]byte(mensagem))
		if err != nil {
			fmt.Println("Erro ao enviar mensagem, desconectando...")
			return
		}
	}
}

func receberMensagens(conexao net.Conn) {
	defer conexao.Close()
	leitor := bufio.NewReader(conexao)
	for {
		mensagem, err := leitor.ReadString('\n')
		if err != nil {
			fmt.Println("Conexão perdida com o servidor.")
			os.Exit(0)
			return
		}
		fmt.Print(mensagem)
	}
}
