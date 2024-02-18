package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/guilhermeabel/rinha-a1-go-api/models"
)

type config struct {
	port int
}

type application struct {
	config     config
	transacoes *models.TransacaoModel
	clientes   *models.ClienteModel
}

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 9999, "API server port")
	flag.Parse()

	app := &application{
		config: cfg,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	_ = srv.ListenAndServe()

}
