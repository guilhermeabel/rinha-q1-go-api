package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/guilhermeabel/rinha-a1-go-api/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	port int
}

type application struct {
	config     config
	db         *pgxpool.Pool
	transacoes *models.TransacaoModel
	clientes   *models.ClienteModel
}

func main() {

	fmt.Printf("Rinha A1 - Go API\n")

	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.Parse()

	dsn := flag.String("dsn", "user=db password=db host=db port=5432 dbname=db", "Postgres data source name")
	flag.Parse()

	ctx := context.Background()

	db, err := openDB(*dsn, ctx)
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

	// go func() {
	// 	for {
	// 		time.Sleep(2 * time.Second)
	// 		err := app.transacoes.LimparTransacoesAntigas(ctx)
	// 		if err != nil {
	// 			fmt.Printf("Erro ao limpar transacoes antigas: %v\n", err)
	// 		}
	// 	}
	// }()

	fmt.Printf("Server running on port %d\n", cfg.port)
	_ = srv.ListenAndServe()
}

func openDB(dsn string, ctx context.Context) (*pgxpool.Pool, error) {
	var maxAttempts = 3
	var db *pgxpool.Pool
	var err error

	for i := 0; i < 10; i++ {
		db, err = pgxpool.New(ctx, dsn)
		if err == nil {
			break
		} else {
			println("Failed to connect to DB, retrying in 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}
	println("Connected to DB")

	db.Config().MaxConnIdleTime = 10 * time.Minute
	db.Config().MaxConnLifetime = 2 * time.Hour
	db.Config().MaxConns = 95
	db.Config().MinConns = 85
	db.Config().HealthCheckPeriod = 10 * time.Minute

	for attempts := 0; attempts < maxAttempts; attempts++ {
		var err error
		if err = db.Ping(ctx); err != nil {
			fmt.Printf("Error pinging database: %s\n", err)
			time.Sleep(time.Second * 5)
		} else {
			break
		}
	}

	return db, nil
}
