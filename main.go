package main

import (
	"flag"
	"fmt"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

const (
	BASE_FEE = 1_000_000 // 0.1 XLM
)

var (
	sequence   int64
	account    string
	memo       string
	privateKey string
)

func init() {
	flag.Int64Var(&sequence, "sequence", 0, "Sequence number to bump to")
	flag.StringVar(&account, "account", "", "Account whos sequence number to bump")
	flag.StringVar(&memo, "memo", "", "Memo to set for the transaction")
	flag.StringVar(&privateKey, "private-key", "", "Private key of the account to use for signing")
}

func main() {
	horizon := horizonclient.DefaultPublicNetClient

	if account == "" {
		fmt.Println("Account to bump must be provided")
	}
	if privateKey == "" {
		fmt.Println("Private key must be provided")
	}

	req := horizonclient.AccountRequest{
		AccountID: account,
	}

	res, err := horizon.AccountDetail(req)
	if err != nil {
		panic(err)
	}

	if sequence <= res.Sequence {
		panic("must bump forward")
	}

	op := txnbuild.BumpSequence{
		BumpTo:        sequence,
		SourceAccount: account,
	}

	params := txnbuild.TransactionParams{
		SourceAccount:        &txnbuild.SimpleAccount{AccountID: account, Sequence: res.Sequence},
		IncrementSequenceNum: true,
		Operations:           []txnbuild.Operation{&op},
		BaseFee:              BASE_FEE,
		Memo:                 txnbuild.MemoText(memo),
		Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(60)}, // 1 minute timeout
	}

	txn, err := txnbuild.NewTransaction(params)
	if err != nil {
		panic(err)
	}

	keypair, err := keypair.ParseFull(privateKey)
	if err != nil {
		panic(err)
	}

	txn, err = txn.Sign(network.PublicNetworkPassphrase, keypair)
	if err != nil {
		panic(err)
	}

	tx, err := horizon.SubmitTransaction(txn)
	if err != nil {
		panic(err)
	}

	fmt.Println(tx)
}
