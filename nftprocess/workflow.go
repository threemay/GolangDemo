package nftprocess

import (
	"internal/pkg/service"
)

type WorkflowInput struct {
	OrderID               string      `json:"orderId"`
	PaymentID             string      `json:"paymentId"`
	ContractID            string      `json:"contractId"`
	OwnershipID           string      `json:"ownershipId"`
	TransactionID         string      `json:"transactionId"`
	TokenID               string      `json:"tokenId"`
	TemplateUrl           string      `json:"templateUrl"`
	ReleaseSymbol         string      `json:"releaseSymbol"`
	ReleaseID             string      `json:"releaseId"`
	ToAddress             string      `json:"toAddress"`
	TransactionType       string      `json:"transactionType"`
	CallbackUrl           string      `json:"callbackUrl"`
	ProductCode           string      `json:"productCode"`
	ExternalTransactionID string      `json:"externalTransactionId"`
	TokenName             string      `json:"tokenName"`
	TokenDescription      string      `json:"tokenDescription"`
	Attributes            []Attribute `json:"attributes"`
}

type Attribute struct {
	TraitType   string      `json:"trait_type,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	DisplayType string      `json:"display_type,omitempty"`
}

type WorkflowOutput struct {
	Status     string `json:"status"`
	TxHash     string `json:"txHash"`
	CurrencyID string `json:"currencyId"`
	UserID     string `json:"userId"`
	TimeStamp  string `json:"timeStamp"`
	ExtraInfo  string `json:"extraInfo"`
}

type Payload struct {
	Input  WorkflowInput  `json:"input"`
	Output WorkflowOutput `json:"output"`
}

const (
	CheckOrder           = "Order"
	CheckPayment         = "Payment"
	Artwork              = "Artwork"
	Mint                 = "Mint"
	Withdraw             = "Withdraw"
	CheckIssuanceBalance = "IssuanceBalance"
	Finalize             = "Finalize"
	Exchange             = "Exchange"
)

type Workflow struct {
	Steps map[string]service.StateMachineStep
}

func ProvideWorkflow(
	order OrderCheckWorkflowStep,
	payment PaymentCheckWorkflowStep,
	artwork ArtworkWorkflowStep,
	mint MintNFTWorkflowStep,
	withdraw WithdrawNFTWorkflowStep,
	issuanceStep PlatformCheckWorkflowStep,
	finalize FinalizeStep,
	exchange ExchangeWorkflowStep,
) Workflow {
	steps := map[string]service.StateMachineStep{
		CheckOrder:           order,
		CheckPayment:         payment,
		Artwork:              artwork,
		Mint:                 mint,
		Withdraw:             withdraw,
		Exchange:             exchange,
		CheckIssuanceBalance: issuanceStep,
		Finalize:             finalize,
	}
	return Workflow{Steps: steps}
}
