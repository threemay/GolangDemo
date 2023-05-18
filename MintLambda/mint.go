package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	workflow "mint/nftprocess"
)

// Check ledger account make sure the balance of the account is sufficient to process the order amount.
// In a way, one state-machine execution is triggered against one order
// receive the message,  set case status to initiated, ready to review.
func handler(ctx context.Context, job workflow.Payload) (workflow.Payload, error) {
	d, err := workflow.Initialize(ctx)
	if err != nil {
		panic("failed initialize domain")
	}
	output, err := d.Steps[workflow.Mint].ProcessJob(ctx, job)
	if err != nil {
		return workflow.Payload{}, err
	}
	return output.(workflow.Payload), nil
}

func main() {
	lambda.Start(handler)
}
