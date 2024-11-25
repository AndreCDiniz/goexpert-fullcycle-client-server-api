package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, "GET", "", nil)
	if err != nil {
		log.Fatal(err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("Erro ao fazer a requisição: %v", err)
		return
	}

	log.Println(response)

	defer response.Body.Close()

}
