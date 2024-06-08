package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CotacaoUSDBRL struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Cot struct {
	Bid string `json:"bid"`
}

func main() {

	http.HandleFunc("/", cotacao)
	http.ListenAndServe(":8080", nil)

}

func cotacao(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 210*time.Millisecond)
	defer cancel()

	cotacao, err := AcessoApiCotacao(ctx)

	if err != nil {
		log.Println("Erro ao consumir API")
		http.Error(w, "Erro ao consumir API", http.StatusRequestTimeout)
		return
	}

	GravarDados(cotacao, ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao.Usdbrl.Bid)
}

func AcessoApiCotacao(ctx context.Context) (*CotacaoUSDBRL, error) {
	select {
	case <-ctx.Done():
		log.Println("Timeout máximo para chamar API ultrapassado")
		return nil, ctx.Err()
	default:
		return GetApiCotacao()
	}
}

func GetApiCotacao() (*CotacaoUSDBRL, error) {
	req, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %w\n", err)
	}
	defer req.Body.Close()
	res, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
	}
	var data CotacaoUSDBRL
	err = json.Unmarshal(res, &data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
	}
	log.Println("Consulta API com sucesso")
	return &data, nil
}

func GravarDados(cotacao *CotacaoUSDBRL, ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("Timeout máximo para persistir dados ultrapassado")
		ctx.Err()
	default:
		//log.Println("Gravacao de dados iniciada")
		sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite-database.db")
		defer sqliteDatabase.Close()
		createTable(sqliteDatabase)
		insertCotacao(sqliteDatabase, cotacao)
		sqliteDatabase.Close()
		log.Println("Gravacao de dados realizado com sucesso")
	}
}

func createTable(db *sql.DB) {
	createCotacaoDolar := `CREATE TABLE IF NOT EXISTS cotacao (
			"idCotacao" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
			"Code"       TEXT,
			"Codein"     TEXT,
			"Name"       TEXT,
			"High"       TEXT,
			"Low"        TEXT,
			"VarBid"     TEXT,
			"PctChange"  TEXT,
			"Bid"        TEXT,
			"Ask"        TEXT,
			"Timestamp"  TEXT,
			"CreateDate" TEXT
		  );`
	//log.Println("Create cotacao table...")
	statement, err := db.Prepare(createCotacaoDolar)
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec()
	//log.Println("cotacao table created")
}

func insertCotacao(db *sql.DB, cotacao *CotacaoUSDBRL) {
	//log.Println("Inserting cotacao record ...")
	insertCotacaoSQL := `INSERT INTO cotacao(		
	Code       ,
	Codein     ,
	Name       ,
	High       ,
	Low        ,
	VarBid     ,
	PctChange  ,
	Bid        ,
	Ask        ,
	Timestamp  ,
	CreateDate ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	statement, err := db.Prepare(insertCotacaoSQL)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(
		cotacao.Usdbrl.Code,
		cotacao.Usdbrl.Codein,
		cotacao.Usdbrl.Name,
		cotacao.Usdbrl.High,
		cotacao.Usdbrl.Low,
		cotacao.Usdbrl.VarBid,
		cotacao.Usdbrl.PctChange,
		cotacao.Usdbrl.Bid,
		cotacao.Usdbrl.Ask,
		cotacao.Usdbrl.Timestamp,
		cotacao.Usdbrl.CreateDate)
	if err != nil {
		log.Fatalln(err.Error())
	}
}
