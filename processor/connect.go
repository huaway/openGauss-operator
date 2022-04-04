package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"

	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "user=root dbname=postgres password=root host=10.111.9.18 port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		select {
		case <-appSignal:
			stop()
		}
	}()

	if err := db.PingContext(ctx); err != nil {
		panic(err)
	}

// "add resource replica_ds_1 (HOST=192.168.202.71, PORT=5432, DB=postgres, USER=gaussdb, PASSWORD=Enmo@123)"
	result, err := db.ExecContext(ctx,
		"add resource replica_ds_1 (HOST=192.168.202.71, PORT=5432, DB=postgres, USER=gaussdb, PASSWORD=Enmo@123);",
	)
	if err != nil {
		fmt.Println(result)
		panic(err)
	}
	fmt.Println(result)

}