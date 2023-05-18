//go:build wireinject
// +build wireinject

package nftprocess

import (
	"context"

	"github.com/google/wire"
	assetgrpc "gitlab.com/tcam-engineering/platform/tcam-serverless/internal/app/asset/grpc"
	ledgergrpc "gitlab.com/tcam-engineering/platform/tcam-serverless/internal/app/ledger/grpc"
	paymentgrpc "gitlab.com/tcam-engineering/platform/tcam-serverless/internal/app/payment/grpc"
	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/app/token/database"
	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/pkg/databaseutil"
	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/pkg/service"
	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/pkg/service/objectstore"
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
	ProvideWorkflow,
	ProvideExchangeStep,
	ProvideMintNftStep,
	ProvideWithdrawNftStep,
	ProvideOrderStep,
	ProvidePaymentStep,
	ProvideArtworkStep,
	ProvideIssuanceStep,
	ProvideFinalizeStep,
)

func Initialize(ctx context.Context) (Workflow, error) {
	wire.Build(WireSet)
	return Workflow{}, nil
}
