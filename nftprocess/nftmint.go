package nftprocess

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/google/uuid"

	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/app/token/database"
	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/pkg/databaseutil"
	"gitlab.com/tcam-engineering/platform/tcam-serverless/internal/pkg/nftutil"

	assetgrpc "gitlab.com/tcam-engineering/platform/tcam-serverless/internal/app/asset/grpc"
)

type MintNFTWorkflowStep struct {
	secretMgr   *secretsmanager.Client
	txnProvider databaseutil.TransactionProvider
	txRepo      database.OnChainTxRepository
	tokenRepo   database.TokenRepository
	assetClient assetgrpc.AssetClient
	ssmClient   *ssm.Client
}

func (s *MintNFTWorkflowStep) AutoMigrate() error {
	tx := s.txnProvider.NoTransaction(context.TODO())
	err := s.txRepo.AutoMigrate(tx)
	if err != nil {
		return fmt.Errorf("failed migrate on-chain tx schema %w", err)
	}
	return nil
}

func (s MintNFTWorkflowStep) mintToken(ctx context.Context, contract *database.TokenContract, tokenID string) (*database.OnChainTransaction, error) {
	switch contract.Protocol {
	case nftutil.ERC1155:
		txDetails, err := nftutil.MintNFT(ctx, &nftutil.ContractInfo{
			Address: contract.Address,
			Name:    contract.Name,
			NodeURI: contract.NodeURI,
		}, s.secretMgr, tokenID, "1", -1)
		if err != nil {
			// TODO wrap error
			return nil, err
		}
		return &database.OnChainTransaction{
			TxHash:     txDetails.TxHash,
			ContractID: contract.ID,
			TokenID:    tokenID,
			Nonce:      txDetails.Nonce,
			GasPrice:   txDetails.GasPrice.String(),
			Status:     database.OnChainTxStatusPending,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported smart contract type %s", contract.Protocol)
	}
}

func (s MintNFTWorkflowStep) ProcessJob(ctx context.Context, job interface{}) (interface{}, error) {
	response, ok := job.(Payload)
	if !ok {
		return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("cannot parse payload"))
	}
	mintJob := response.Input
	db := s.txnProvider.NoTransaction(ctx)
	contract := &database.TokenContract{}
	if mintJob.ContractID == "" {
		contract, err := s.tokenRepo.FindOne(db, map[string]interface{}{
			"currency_id": response.Output.CurrencyID,
		})
		if err != nil {
			return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("cannot find token contract by currency id %s", err))
		}
		mintJob.ContractID = contract.ID.String()
		response.Input.ContractID = contract.ID.String()
	} else {
		cId, err := uuid.Parse(mintJob.ContractID)
		if err != nil {
			return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("cannot parse ContractID %s", err))
		}
		contract, err = s.tokenRepo.FindByID(db, cId)
		if err != nil {
			return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("failed to find token contract %s", err))
		}
	}
	tx, err := s.mintToken(ctx, contract, mintJob.TokenID)
	if err != nil {
		return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("failed to mint NFT token %s", err))
	}
	if tx.TxHash == "" {
		return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("on chain transaction is not created %s", err))
	}
	if mintJob.TransactionID != "" {
		_, err = s.assetClient.PatchNftTransactions(ctx, &assetgrpc.NftTransaction{
			Id:           mintJob.TransactionID,
			TxHash:       tx.TxHash,
			Nonce:        strconv.FormatUint(tx.Nonce, 10),
			GasFee:       tx.GasPrice,
			UserId:       response.Output.UserID,
			PaymentId:    mintJob.PaymentID,
			Chain:        contract.Chain,
			ChainNetwork: contract.ChainNetwork,
			ToAddress:    contract.Address,
		})
		if err != nil {
			return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("failed to update NftTransactions %s", err))
		}
	} else {
		nftTx, err := s.assetClient.PostNftTransactions(ctx, &assetgrpc.NftTransaction{
			TxHash: tx.TxHash,
			Nonce:  strconv.FormatUint(tx.Nonce, 10),
			GasFee: tx.GasPrice,
			TokenInfo: []*assetgrpc.TokenInfo{{
				TokenId: mintJob.TokenID,
				Amount:  "1",
			}},
			ContractId:   mintJob.ContractID,
			ToAddress:    contract.Address,
			PaymentId:    mintJob.PaymentID,
			TxType:       mintJob.TransactionType,
			UserId:       response.Output.UserID,
			Chain:        contract.Chain,
			ChainNetwork: contract.ChainNetwork,
		})
		if err != nil {
			return processCallback(ctx, &response, s.ssmClient, s.assetClient, fmt.Sprintf("failed to create nft transactions %s", err))
		}
		response.Input.TransactionID = nftTx.Id
	}
	response.Output.TxHash = tx.TxHash
	response.Output.CurrencyID = contract.CurrencyID.String()
	response.Output.TimeStamp = ""
	return response, nil
}

func ProvideMintNftStep(
	txnProvider databaseutil.TransactionProvider,
	secretMgr *secretsmanager.Client,
	tokenRepo database.TokenRepository,
	txRepo database.OnChainTxRepository,
	assetClient assetgrpc.AssetClient,
	ssmClient *ssm.Client,
) MintNFTWorkflowStep {
	return MintNFTWorkflowStep{
		secretMgr,
		txnProvider,
		txRepo,
		tokenRepo,
		assetClient,
		ssmClient,
	}
}
