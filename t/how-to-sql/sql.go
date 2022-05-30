package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	do(context.TODO())
}

func do(ctx context.Context) {

	db, err := sqlx.Open("pgx", "postgres://ivasw:ivasw@127.0.0.1:5432/ivasw")
	if err != nil {
		log.Fatalln("open", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(3)

	conn, err := db.Connx(ctx)
	if err != nil {
		log.Fatalln("ConnX", err)
	}
	defer conn.Close()

	stmt1, err := db.PrepareNamedContext(ctx, "INSERT INTO protocol(code,name) values(:code0,:name0)")
	if err != nil {
		log.Fatalln("PrepareNamedContext", 1, err)
	}
	defer stmt1.Close()

	stmt2, err := db.PrepareNamedContext(ctx, "INSERT INTO protocol(code,name) values(:code0,:name0)")
	if err != nil {
		log.Fatalln("PrepareNamedContext", 2, err)
	}
	defer stmt2.Close()

	tx1, err := db.BeginTxx(ctx, nil)
	if err != nil {
		log.Fatalln("BeginTxx", 1, err)
	}

	_, err = tx1.NamedStmtContext(ctx, stmt1).ExecContext(ctx, map[string]interface{}{"code0": "asd", "name0": "ASD"})
	if err != nil {
		log.Fatalln("NamedStmtContext.ExecContext", 1, err)
	}

	tx2, err := db.BeginTxx(ctx, nil)
	if err != nil {
		log.Fatalln("BeginTxx", 2, err)
	}

	_, err = tx2.NamedStmtContext(ctx, stmt1).ExecContext(ctx, map[string]interface{}{"code0": "qwe", "name0": "QWE"})
	if err != nil {
		log.Fatalln("NamedStmtContext.ExecContext", 2, 1, err)
	}
	_, err = tx2.NamedStmtContext(ctx, stmt1).ExecContext(ctx, map[string]interface{}{"code0": "wer", "name0": "WER"})
	if err != nil {
		log.Fatalln("NamedStmtContext.ExecContext", 2, 2, err)
	}
	_, err = tx2.NamedStmtContext(ctx, stmt1).ExecContext(ctx, map[string]interface{}{"code0": "ert", "name0": "ERT"})
	if err != nil {
		log.Fatalln("NamedStmtContext.ExecContext", 2, 3, err)
	}

	err = tx1.Commit()
	if err != nil {
		log.Fatalln("Commit", 1, err)
	}

	err = tx2.Commit()
	if err != nil {
		log.Fatalln("Commit", 2, err)
	}

	var wg, bg sync.WaitGroup
	bg.Add(5)
	wg.Add(5)

	for j := 0; j < 5; j++ {
		go run(ctx, stmt1, j, &wg, &bg)
	}

	fmt.Println("Press Ctrl+С")

	sig, done := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig.Done():
		done()
	}

	_, err = conn.ExecContext(ctx, "DELETE FROM protocol WHERE code not in ($1,$2)", "sip", "h323")
	if err != nil {
		log.Fatalln("ExecContext", err)
	}

}

func run(ctx context.Context, stmt *sqlx.NamedStmt, j int, wg, bg *sync.WaitGroup) {
	defer wg.Done()
	bg.Done()
	bg.Wait()
	for i := 0; i < 99; i++ {
		x := fmt.Sprintf("%d%02d", j, i)
		_, err := stmt.ExecContext(ctx, map[string]interface{}{
			"code0": x,
			"name0": x,
		})
		if err != nil {
			log.Fatalln("ExecContext", i, err)
		}
	}
}
