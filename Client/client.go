package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type exchangeReq struct {
	Bid string `json:"bid"`
}

func main() {

	file, err := os.OpenFile("client.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Println("Log inicializado")

	exchange, err := geExchange()
	if err != nil {
		log.Fatal(err)
	}

	err = addRecord(exchange)
	if err != nil {
		log.Fatal(err)
	}

	ex, _ := json.Marshal(exchange)
	log.Print("Requisição realizada com sucesso:")
	log.Println(string(ex))
}

func addRecord(exch *exchangeReq) error {

	file, err := os.OpenFile("cotacao.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	rec := []byte("Dólar: " + string(exch.Bid) + "\n")
	_, err = file.Write(rec)
	if err != nil {
		return err
	}

	return nil
}

func geExchange() (*exchangeReq, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c exchangeReq
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
