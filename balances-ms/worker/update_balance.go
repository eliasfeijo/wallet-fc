package worker

import (
	"database/sql"
	"encoding/json"
	"log"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/eliasfeijo/wallet-fc/balances-ms/database/dao"
	"github.com/eliasfeijo/wallet-fc/balances-ms/database/model"
	"github.com/eliasfeijo/wallet-fc/balances-ms/pkg/kafka"
)

type UpdateBalancePayload struct {
	AccountIDFrom        string  `json:"account_id_from"`
	AccountIDTo          string  `json:"account_id_to"`
	BalanceAccountIDFrom float64 `json:"balance_account_id_from"`
	BalanceAccountIDTo   float64 `json:"balance_account_id_to"`
}

type UpdateBalanceWorker struct {
	db       *sql.DB
	consumer *kafka.Consumer
	dao      *dao.AccountBalanceDAO
}

func NewUpdateBalanceWorker(db *sql.DB) *UpdateBalanceWorker {
	consumer := kafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "wallet",
	}, []string{"balances"})
	dao := dao.NewAccountBalanceDAO()
	return &UpdateBalanceWorker{db, consumer, dao}
}

func (u *UpdateBalanceWorker) Work() {
	messages := make(chan *ckafka.Message)
	go u.consumer.Consume(messages)
	for {
		msg := <-messages
		payload := UpdateBalancePayload{}
		err := json.Unmarshal(msg.Value, &payload)
		if err != nil {
			log.Printf("error unmarshalling message: %v", err)
			continue
		}
		from := model.AccountBalance{
			AccountID: payload.AccountIDFrom,
			Balance:   payload.BalanceAccountIDFrom,
			DateTime:  msg.Timestamp,
		}
		to := model.AccountBalance{
			AccountID: payload.AccountIDTo,
			Balance:   payload.BalanceAccountIDTo,
			DateTime:  msg.Timestamp,
		}
		err = u.updateBalance(from, to)
		if err != nil {
			log.Printf("error updating balance: %v", err)
		}
	}
}

func (u *UpdateBalanceWorker) updateBalance(from model.AccountBalance, to model.AccountBalance) error {
	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	err = u.dao.Save(tx, &from)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = u.dao.Save(tx, &to)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
