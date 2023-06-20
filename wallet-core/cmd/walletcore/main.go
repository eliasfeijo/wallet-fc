package main

import (
	"context"
	"database/sql"
	"fmt"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/database"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/entity"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/event"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/event/handler"
	createaccount "github.com/eliasfeijo/fc-wallet/wallet-core/internal/usecase/create_account"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/usecase/create_client"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/usecase/create_transaction"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/web"
	"github.com/eliasfeijo/fc-wallet/wallet-core/internal/web/webserver"
	"github.com/eliasfeijo/fc-wallet/wallet-core/pkg/events"
	"github.com/eliasfeijo/fc-wallet/wallet-core/pkg/kafka"
	"github.com/eliasfeijo/fc-wallet/wallet-core/pkg/uow"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", "root", "root", "mysql", "3306", "wallet"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Exec("Create table if not exists clients (id varchar(255), name varchar(255), email varchar(255), created_at date)")
	db.Exec("Create table if not exists accounts (id varchar(255), client_id varchar(255), balance int, created_at date)")
	db.Exec("Create table if not exists transactions (id varchar(255), account_id_from varchar(255), account_id_to varchar(255), amount int, created_at date)")

	configMap := ckafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "wallet",
	}
	kafkaProducer := kafka.NewKafkaProducer(&configMap)

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("TransactionCreated", handler.NewTransactionCreatedKafkaHandler(kafkaProducer))
	eventDispatcher.Register("BalanceUpdated", handler.NewUpdateBalanceKafkaHandler(kafkaProducer))
	transactionCreatedEvent := event.NewTransactionCreated()
	balanceUpdatedEvent := event.NewBalanceUpdated()

	clientDb := database.NewClientDB(db)
	accountDb := database.NewAccountDB(db)

	rows, _ := db.Query("SELECT 1 FROM clients LIMIT 1")
	if !rows.Next() {
		seedDb(clientDb, accountDb)
	}

	ctx := context.Background()
	uow := uow.NewUow(ctx, db)

	uow.Register("AccountDB", func(tx *sql.Tx) interface{} {
		return database.NewAccountDB(db)
	})

	uow.Register("TransactionDB", func(tx *sql.Tx) interface{} {
		return database.NewTransactionDB(db)
	})
	createTransactionUseCase := create_transaction.NewCreateTransactionUseCase(uow, eventDispatcher, transactionCreatedEvent, balanceUpdatedEvent)
	createClientUseCase := create_client.NewCreateClientUseCase(clientDb)
	createAccountUseCase := createaccount.NewCreateAccountUseCase(accountDb, clientDb)

	webserver := webserver.NewWebServer(":8080")

	clientHandler := web.NewWebClientHandler(*createClientUseCase)
	accountHandler := web.NewWebAccountHandler(*createAccountUseCase)
	transactionHandler := web.NewWebTransactionHandler(*createTransactionUseCase)

	webserver.AddHandler("/clients", clientHandler.CreateClient)
	webserver.AddHandler("/accounts", accountHandler.CreateAccount)
	webserver.AddHandler("/transactions", transactionHandler.CreateTransaction)

	fmt.Println("Server is running")
	webserver.Start()
}

func seedDb(clientDb *database.ClientDB, accountDb *database.AccountDB) {
	client1, _ := entity.NewClient("John", "j@j.com")
	clientDb.Save(client1)
	client2, _ := entity.NewClient("Mary", "m@m.com")
	clientDb.Save(client2)
	account1 := entity.NewAccount(client1)
	accountDb.Save(account1)
	account2 := entity.NewAccount(client2)
	accountDb.Save(account2)
}
