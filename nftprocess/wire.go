//go:build wireinject
// +build wireinject

package nftprocess

import (
	"context"

	assetgrpc "internal/app/asset/grpc"
	ledgergrpc "internal/app/ledger/grpc"
	paymentgrpc "internal/app/payment/grpc"
	"internal/app/token/database"
	"internal/pkg/databaseutil"
	"internal/pkg/service"
	"internal/pkg/service/objectstore"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	databaseutil.NewTransactionProvider,
	database.NewTokenRepository,
	database.NewOnChainTxRepository,
	service.ProvideGormDB,
	service.ProvidePostgresConfig,
	service.AWSWireSet,
	objectstore.ProvideObjectStore,
	paymentgrpc.ProvidePaymentClient,
	assetgrpc.ProvideAssetClient,
	ledgergrpc.ProvideLedgerClient,
)

func Initialize(ctx context.Context) (Workflow, error) {
	wire.Build(WireSet)
	return Workflow{}, nil
}
