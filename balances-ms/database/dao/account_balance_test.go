package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/eliasfeijo/wallet-fc/balances-ms/database/model"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AccountBalanceDAOTestSuite struct {
	suite.Suite
	db *sql.DB
}

func (suite *AccountBalanceDAOTestSuite) SetupTest() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	stmt, err := db.Prepare("CREATE TABLE account_balances (account_id VARCHAR(255) PRIMARY KEY, balance INT, date_time DATETIME)")
	if err != nil {
		panic(err)
	}
	_, err = stmt.Exec()
	if err != nil {
		panic(err)
	}
	suite.db = db
}

func (suite *AccountBalanceDAOTestSuite) TearDownTest() {
	suite.db.Exec("DROP TABLE account_balances")
	suite.db.Close()
}

func TestAccountBalanceDAOSuite(t *testing.T) {
	suite.Run(t, new(AccountBalanceDAOTestSuite))
}

func (suite *AccountBalanceDAOTestSuite) InsertAccountBalance(tx *sql.Tx, ab *model.AccountBalance) error {
	stmt, err := tx.Prepare("INSERT INTO account_balances (account_id, balance, date_time) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(ab.AccountID, ab.Balance, ab.DateTime)
	if err != nil {
		return err
	}
	return nil
}

func (suite *AccountBalanceDAOTestSuite) TestFindAccountByID() {
	tx, err := suite.db.Begin()
	assert.Nil(suite.T(), err)
	dao := NewAccountBalanceDAO()
	_, err = dao.FindByAccountID(tx, "1")
	assert.EqualError(suite.T(), err, "sql: no rows in result set")
	ab := model.NewAccountBalance("1", 100, time.Now())
	err = suite.InsertAccountBalance(tx, ab)
	assert.Nil(suite.T(), err)
	ab, err = dao.FindByAccountID(tx, "1")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "1", ab.AccountID)
	assert.Equal(suite.T(), 100.0, ab.Balance)
}

func (suite *AccountBalanceDAOTestSuite) TestSave() {
	tx, err := suite.db.Begin()
	assert.Nil(suite.T(), err)
	ab := model.NewAccountBalance("1", 100, time.Now())
	dao := NewAccountBalanceDAO()
	err = dao.Save(tx, ab)
	assert.Nil(suite.T(), err)
	ab, err = dao.FindByAccountID(tx, "1")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "1", ab.AccountID)
	assert.Equal(suite.T(), 100.0, ab.Balance)
	ab = model.NewAccountBalance("1", 200, time.Now())
	err = dao.Save(tx, ab)
	assert.Nil(suite.T(), err)
	ab, err = dao.FindByAccountID(tx, "1")
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "1", ab.AccountID)
	assert.Equal(suite.T(), 200.0, ab.Balance)
	ab = model.NewAccountBalance("1", 150, time.Now().Add(-1*time.Hour))
	err = dao.Save(tx, ab)
	assert.EqualError(suite.T(), err, "newer balance already exists")
}
