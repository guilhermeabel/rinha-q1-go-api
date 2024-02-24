package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/guilhermeabel/rinha-a1-go-api/models"
)

type config struct {
	port int
}

type application struct {
	config     config
	db         *sql.DB
	transacoes *models.TransacaoModel
	clientes   *models.ClienteModel
}

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 9999, "API server port")
	flag.Parse()

	dsn := flag.String("dsn", "root:@/rinha?parseTime=true", "MySQL data source name")
	flag.Parse()

	db, err := openDB(*dsn)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	app := &application{
		config:     cfg,
		db:         db,
		transacoes: &models.TransacaoModel{DB: db},
		clientes:   &models.ClienteModel{DB: db},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Printf("Server running on port %d\n", cfg.port)
	_ = srv.ListenAndServe()
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
