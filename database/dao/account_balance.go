package dao

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/eliasfeijo/wallet-consumer-golang/database/model"
)

type AccountBalanceDAO struct {
	db *sql.DB
}

func NewAccountBalanceDAO(db *sql.DB) *AccountBalanceDAO {
	return &AccountBalanceDAO{db: db}
}

func (a *AccountBalanceDAO) FindByAccountID(id string) (*model.AccountBalance, error) {
	var balance float64
	var dateTime time.Time

	stmt, err := a.db.Prepare("SELECT balance, date_time FROM account_balances WHERE account_id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	err = row.Scan(&balance, &dateTime)
	if err != nil {
		return nil, err
	}
	return model.NewAccountBalance(id, balance, dateTime), nil
}

func (a *AccountBalanceDAO) Save(m *model.AccountBalance) error {
	existing, err := a.FindByAccountID(m.AccountID)
	if err != nil {
		return a.insert(m)
	}
	if m.DateTime.Before(existing.DateTime) {
		return fmt.Errorf("newer balance already exists")
	}
	return a.update(m)
}

func (a *AccountBalanceDAO) insert(m *model.AccountBalance) error {
	stmt, err := a.db.Prepare("INSERT INTO account_balances (account_id, balance, date_time) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.AccountID, m.Balance, m.DateTime)
	if err != nil {
		return err
	}
	return nil
}

func (a *AccountBalanceDAO) update(m *model.AccountBalance) error {
	stmt, err := a.db.Prepare("UPDATE account_balances SET balance = ?, date_time = ? WHERE account_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.Balance, m.DateTime, m.AccountID)
	if err != nil {
		return err
	}
	return nil
}
