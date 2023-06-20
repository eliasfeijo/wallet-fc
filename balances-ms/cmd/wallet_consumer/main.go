package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eliasfeijo/wallet-fc/balances-ms/database/dao"
	"github.com/eliasfeijo/wallet-fc/balances-ms/database/model"
	"github.com/eliasfeijo/wallet-fc/balances-ms/worker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type AccountBalanceDTO struct {
	Balance float64 `json:"balance"`
}

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", "root", "root", "mysql2", "3306", "balances"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("CREATE TABLE IF NOT EXISTS account_balances (account_id VARCHAR(255) PRIMARY KEY, balance INT, date_time DATETIME)")
	rows, _ := db.Query("SELECT 1 FROM account_balances LIMIT 1")
	if !rows.Next() {
		seedDB(db)
	}

	dao := dao.NewAccountBalanceDAO()

	balanceWorker := worker.NewUpdateBalanceWorker(db)
	go balanceWorker.Work()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/balances/{id}", getBalance(db, dao))

	log.Println("Listening on port 3003")
	http.ListenAndServe(":3003", r)
}

func getBalance(db *sql.DB, dao *dao.AccountBalanceDAO) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountId := chi.URLParam(r, "id")
		tx, err := db.Begin()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error getting balance"))
			return
		}
		ab, err := dao.FindByAccountID(tx, accountId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error getting balance"))
			return
		}
		dto := AccountBalanceDTO{Balance: ab.Balance}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto)
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
