package main

import (
	"context"
	"flag"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"vav-tech.ru/snippetbox/pkg/models/postgres"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *postgres.SnippetModel
}

func main() {
	addr := flag.String("addr", ":4000", "Сетевой адресс HTTP")
	dsn := flag.String("dsn", "postgresql://postgres:postgres@localhost/snippetbox", "URL бызы данных")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	dbPool, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer dbPool.Close()

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		snippets: &postgres.SnippetModel{DB: dbPool},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Start server on: %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}
	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			return nil, err
		}
	}
	return f, nil
}

func openDB(dbUrl string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		return nil, err
	}
	if err = dbpool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return dbpool, nil
}
