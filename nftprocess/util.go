package nftprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	assetgrpc "internal/app/asset/grpc"
	ledgergrpc "internal/app/ledger/grpc"
	"internal/pkg/configutil"
	"internal/pkg/cryptoutil"
	"internal/pkg/domainutil"
	"internal/pkg/service/http"
)

func checkOrderExpiry(order assetgrpc.AssetOrder) (bool, error) {
	if order.GetExpireAt() == "" {
		// return nil error but set the status to StatusError will end the state machine and leave order in a
		// blocked status, manual operation is required to process the order
		return false, domainutil.ErrInternalServerError(fmt.Errorf("nil order expiry time "), "")
	}
	expireAt, err := time.Parse("2006-01-02T15:04:05Z07:00", order.ExpireAt)
	if err != nil {
		return false, domainutil.ErrInternalServerError(fmt.Errorf("invalid order expiry time "), "")
	}
	now := time.Now()
	if expireAt.Before(now) {
		return true, nil
	}
	return false, nil
}

func checkTimeStamp(timeStamp string) (bool, error) {
	expireAt, err := time.Parse("2006-01-02T15:04:05Z07:00", timeStamp)
	if err != nil {
		return false, domainutil.ErrInternalServerError(fmt.Errorf("invalid timeStamp time "), "")
	}
	now := time.Now()
	if expireAt.Before(now) {
		return true, nil
	}
	return false, nil
}

func checkAccountBalance(ctx context.Context, ledgerClient ledgergrpc.LedgerClient, currency, holder string, category *ledgergrpc.AccountCategory, amt decimal.Decimal) (bool, error) {
	checkAccReq := ledgergrpc.GetLedgerAccountsRequest{
		HolderIds:  []string{holder},
		Currencies: []string{currency},
	}
	if category != nil {
		checkAccReq.Categories = []ledgergrpc.AccountCategory{*category}
	}
	accResp, err := ledgerClient.GetLedgerAccounts(ctx, &checkAccReq)
	if err != nil {
		return false, err
	}
	accList := accResp.Items
	if len(accList) == 0 {
		return false, fmt.Errorf("failed to find ledger accounts for %s", holder)
	}
	holderAcc := accList[0]
	balance, err := decimal.NewFromString(holderAcc.GetBalance())
	if err != nil {
		return false, domainutil.ErrInternalServerError(err, "failed to new decimal from string")
	}
	lockedBal, err := decimal.NewFromString(holderAcc.GetLockedBalance())
	if err != nil {
		lockedBal = decimal.Zero
		log.Errorf("failed to parse locked balance for account: %s", holder)
		return false, domainutil.ErrInternalServerError(err, "failed to new decimal from string")
	}
	log.Infof("account %s balance: %s locked balance: %s amount: %s", holderAcc.GetId(), balance.String(), lockedBal.String(), amt.String())
	ret := balance.Add(lockedBal).Add(amt)
	return !ret.IsPositive(), nil
}

func processCallback(ctx context.Context, workflowPayload *Payload, ssmClient *ssm.Client, assetClient assetgrpc.AssetClient, errMsg string) (interface{}, error) {
	if errMsg != "" {
		workflowPayload.Output.Status = StatusError
		workflowPayload.Output.ExtraInfo = workflowPayload.Output.ExtraInfo + errMsg
	}
	payload := NftSuccessNotification{}
	tx, err := assetClient.GetNftTransactionsById(ctx, &assetgrpc.NftTransaction{
		Id: workflowPayload.Input.TransactionID,
	})
	if err != nil {
		workflowPayload.Output.ExtraInfo = workflowPayload.Output.ExtraInfo + fmt.Sprintf("err for GetNftTransactionsById: %s", err)
		payload.Status = NftSuccessNotificationStatusFailed
		payload.PaymentID = workflowPayload.Input.PaymentID
		payload.TokenID = workflowPayload.Input.TokenID
		payload.ExternalTransactionID = workflowPayload.Input.ExternalTransactionID
		errSignAndSend := signAndSend(ctx, payload, workflowPayload.Input.CallbackUrl, ssmClient)
		return *workflowPayload, fmt.Errorf("errMsg: %s, processCallback err: %s, GetNftTransactionsById err: %s", errMsg, errSignAndSend, err)
	}
	payload = NftSuccessNotification{
		Chain:                 tx.Chain,
		ChainNetwork:          tx.ChainNetwork,
		ReleaseSymbol:         workflowPayload.Input.ReleaseSymbol,
		ReleaseID:             workflowPayload.Input.ReleaseID,
		TokenID:               workflowPayload.Input.TokenID,
		TxHash:                tx.TxHash,
		GasFee:                tx.GasFee,
		ToAddress:             workflowPayload.Input.ToAddress,
		UserID:                workflowPayload.Output.UserID,
		PaymentID:             workflowPayload.Input.PaymentID,
		ExternalTransactionID: workflowPayload.Input.ExternalTransactionID,
	}
	if workflowPayload.Output.Status != StatusSufficient {
		workflowPayload.Output.ExtraInfo = workflowPayload.Output.ExtraInfo + ", not sufficient"
		payload.Status = NftSuccessNotificationStatusFailed
	}
	if workflowPayload.Input.TransactionType == "Mint" {
		payload.UserID = ""
	}
	if errMsg == "" {
		return *workflowPayload, signAndSend(ctx, payload, workflowPayload.Input.CallbackUrl, ssmClient)
	} else {
		return *workflowPayload,
			fmt.Errorf("errMsg: %s, signAndSend err: %s", errMsg, signAndSend(ctx, payload, workflowPayload.Input.CallbackUrl, ssmClient))
	}
}

func signAndSend(ctx context.Context, payload NftSuccessNotification, callbackUrl string, ssmClient *ssm.Client) error {
	env := os.Getenv("ENV")
	domain := "mercury"
	key, err := configutil.GetParameter(ctx, ssmClient, domain, env, "signature_key")
	if err != nil {
		return err
	}
	var signatureKey SignatureKey
	err = json.Unmarshal([]byte(key), &signatureKey)
	if err != nil {
		return err
	}
	messageBytes, err := json.Marshal(payload)
	signature, err := cryptoutil.Sign(signatureKey.PrivateKey, messageBytes)
	if err != nil {
		return err
	}
	payload.Signature = signature
	messageBytes, err = json.Marshal(payload)
	if err != nil {
		return err
	}
	// send request to callback URL
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	_, err = http.DoHttpRequest(messageBytes, "POST", callbackUrl, headers, nil)
	if err != nil {
		return err
	}
	return nil
}
