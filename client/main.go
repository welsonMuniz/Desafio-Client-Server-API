package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080", nil)
	if err != nil {
		log.Println("Timeout máximo para acesso Server ultrapassado 1")
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Timeout máximo para acesso Server ultrapassado 2")
		panic(err)
	}

	defer res.Body.Close()

	select {
	case <-ctx.Done():
		// The context deadline has been exceeded.
		log.Println("Timeout máximo para acesso Server ultrapassado 3")
		log.Println(ctx.Err())

	default:
		resp, err := io.ReadAll((res.Body))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
		}

		//gravar arquivo
		file, err := os.Create("cotacao.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
		}
		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("Dólar: %s", resp))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao gravar arquivo: %v\n", err)
		}

	}
}
