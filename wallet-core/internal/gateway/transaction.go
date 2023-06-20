package gateway

import "github.com/eliasfeijo/fc-wallet/wallet-core/internal/entity"

type TransactionGateway interface {
	Create(transaction *entity.Transaction) error
}
