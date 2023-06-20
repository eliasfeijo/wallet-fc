package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eliasfeijo/wallet-consumer-golang/database/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", "root", "root", "localhost", "3307", "wallet_consumer"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("CREATE TABLE IF NOT EXISTS account_balances (account_id VARCHAR(255) PRIMARY KEY, balance INT, date_time DATETIME)")
	rows, _ := db.Query("SELECT 1 FROM account_balances LIMIT 1")
	if !rows.Next() {
		seedDB(db)
	}
}

func seedDB(db *sql.DB) {
	ab := model.NewAccountBalance(uuid.New().String(), 0, time.Now())
	err := insertAccountBalance(db, ab)
	if err != nil {
		panic(err)
	}
	ab = model.NewAccountBalance(uuid.New().String(), 1000, time.Now())
	err = insertAccountBalance(db, ab)
	if err != nil {
		panic(err)
	}
}

func insertAccountBalance(db *sql.DB, ab *model.AccountBalance) error {
	stmt, err := db.Prepare("INSERT INTO account_balances (account_id, balance, date_time) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(ab.AccountID, ab.Balance, ab.DateTime)
	if err != nil {
		return err
	}
	return nil
}
