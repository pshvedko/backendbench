package main

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func do(ctx context.Context) {

	db, err := sqlx.Open("pgx", "postgres://ivasw:ivasw@127.0.0.1:5432/ivasw")
	if err != nil {
		log.Fatalln("open", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(33)

	conn1, err := db.Connx(ctx)
	if err != nil {
		log.Fatalln("ConnX", err)
	}
	defer conn1.Close()

	defer func() {
		_, err = conn1.ExecContext(ctx, "DELETE FROM protocol WHERE code not in ($1,$2)", "sip", "h323")
		if err != nil {
			log.Fatalln("ExecContext", err)
		}
	}()

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

	stmtX, err := tx2.PrepareNamedContext(ctx, "INSERT INTO protocol(code,name) values(:code0,:name0)")
	if err != nil {
		log.Fatalln("PrepareNamedContext", 3, err)
	}
	defer stmtX.Close()

	_, err = stmtX.ExecContext(ctx, map[string]interface{}{"code0": "xxx", "name0": "XXX"})
	if err != nil {
		log.Fatalln("NamedStmtContext.ExecContext", 3, 1, err)
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
		go run(ctx, stmt2, j, &wg, &bg)
	}

	wg.Wait()

	_, err = stmtX.ExecContext(ctx, map[string]interface{}{"code0": "zzz", "name0": "ZZZ"})
	if err != nil {
		log.Println("NamedStmtContext.ExecContext", 3, 2, err)
	}

	db.SetMaxOpenConns(1 + 1) // conn1 + stmt3

	stmt3, err := db.PrepareNamedContext(ctx, "SELECT * FROM protocol LIMIT 10")
	if err != nil {
		log.Fatalln("PrepareNamedContext", 3, err)
	}

	bg.Add(2)
	wg.Add(2)

	go row(ctx, stmt3, 0, &wg, &bg)
	go row(ctx, stmt3, 1, &wg, &bg)

	wg.Wait()

	fmt.Println("Press Ctrl+С")

	sig, done := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sig.Done():
		done()
	}
}

func row(ctx context.Context, stmt *sqlx.NamedStmt, j int, wg, bg *sync.WaitGroup) {
	defer wg.Done()
	bg.Done()
	bg.Wait()
	log.Println("===>row(", j, ")")
	rows, err := stmt.QueryContext(ctx, map[string]interface{}{})
	if err != nil {
		log.Fatalln("QueryContext", j, err)
	}
	for rows.Next() {
		time.Sleep(time.Second / 2)
		var a, b, c, d string
		err = rows.Scan(&a, &b, &c, &d)
		if err != nil {
			log.Fatalln("Scan", j, err)
		}
		log.Println(j, a, b, c, d)
	}
	err = rows.Close()
	if err != nil {
		log.Fatalln("Close", j, err)
	}
}

func run(ctx context.Context, stmt *sqlx.NamedStmt, j int, wg, bg *sync.WaitGroup) {
	defer wg.Done()
	bg.Done()
	bg.Wait()
	log.Println("===>run(", j, ")")
	for i := 0; i < 99; i++ {
		x := fmt.Sprintf("%d%02d", j, i)
		_, err := stmt.ExecContext(ctx, map[string]interface{}{
			"code0": x,
			"name0": x,
		})
		if err != nil {
			log.Fatalln("ExecContext", j, i, err)
		}
	}
}

func main() {
	do(context.TODO())
}
