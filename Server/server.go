package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type exchangeReqDTO struct {
	Code        string `json:"code"`
	Codein      string `json:"codein"`
	Name        string `json:"name"`
	High        string `json:"high"`
	Low         string `json:"low"`
	VarBid      string `json:"varBid"`
	PctChange   string `json:"pctChange"`
	Bid         string `json:"bid"`
	Ask         string `json:"ask"`
	Timestamp   string `json:"timestamp"`
	Create_date string `json:"create_date"`
	gorm.Model
}

type exchangeUsdbrl struct {
	Usdbrl exchangeReqDTO `json:"usdbrl"`
}

type exchangeRespDTO struct {
	Bid string `json:"bid"`
}

func main() {
	file, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Println("Log inicializado")

	http.HandleFunc("/cotacao", cotacaoHandler)

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func cotacaoHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("request /cotacao: iniciada")
	defer log.Println("request /cotacao: finalizada")

	exchange, err := getDolarToReal()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Println("request /cotacao: erro na chamda externa")
		return
	}

	err = insertExchange(exchange)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Println("request /cotacao: erro na chamada do banco de dados")
		return
	}

	var resp exchangeRespDTO
	resp.Bid = exchange.Usdbrl.Bid

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	log.Println("request /cotacao: status code 200")
}

func insertExchange(exch *exchangeUsdbrl) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// inicialização do banco de dados
	db, err := gorm.Open(sqlite.Open("server.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&exchangeReqDTO{})
	if err != nil {
		return err
	}

	result := db.WithContext(ctx).Create(&exch.Usdbrl)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Recupera as informações da cotação do Dolar para o Real
func getDolarToReal() (*exchangeUsdbrl, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
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

	var c exchangeUsdbrl
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
