package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type CurrencyResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {

	db, err := sql.Open("sqlite", "cotacao.db")
	if err != nil {
		log.Fatalf("Erro ao conectar com o banco de dados: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
    		id INTEGER PRIMARY KEY AUTOINCREMENT,
    		bid TEXT,
    		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctxApi, cancelApi := context.WithTimeout(context.Background(), time.Millisecond*2000)
		defer cancelApi()

		req, err := http.NewRequestWithContext(ctxApi, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			log.Printf("Erro ao criar request: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Erro ao fazer a requisição: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var CurrencyResponse CurrencyResponse
		if err := json.NewDecoder(resp.Body).Decode(&CurrencyResponse); err != nil {
			log.Printf("Erro ao decodificar a requisição: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctxDb, cancelDb := context.WithTimeout(context.Background(), time.Millisecond*10)
		defer cancelDb()

		_, err = db.ExecContext(ctxDb, "INSERT INTO cotacoes (bid) VALUES (?)", CurrencyResponse.USDBRL.Bid)
		if err != nil {
			log.Printf("Erro ao inserir no banco de dados: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": CurrencyResponse.USDBRL.Bid})
	})

	fmt.Println("Server is running at port 8282")
	log.Fatal(http.ListenAndServe(":8282", nil))

}
