package rosetta

import (
	"strconv"

	v1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	signing2 "cosmossdk.io/x/tx/signing"
	"google.golang.org/protobuf/types/known/anypb"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	crgerrs "github.com/cosmos/rosetta/lib/errors"
)

func parseSignerData(signerData authsigning.SignerData) signing2.SignerData {
	parsedSignerDataPublicKey := anypb.Any{
		TypeUrl: sdk.MsgTypeURL(signerData.PubKey),
		Value:   signerData.PubKey.Bytes(),
	}
	return signing2.SignerData{Address: strconv.FormatUint(signerData.AccountNumber, 10), ChainID: signerData.ChainID, AccountNumber: signerData.AccountNumber, Sequence: signerData.Sequence, PubKey: &parsedSignerDataPublicKey}
}

func parseTxTip(tx authsigning.Tx) txv1beta1.Tip {
	parsedTipAmount := []*v1beta1.Coin{}
	tipper := string(tx.FeePayer())

	if tx.GetTip() != nil {
		for _, txCoin := range tx.GetTip().Amount {
			parsedTipAmount = append(parsedTipAmount, &v1beta1.Coin{
				Denom:  txCoin.Denom,
				Amount: txCoin.Amount.String(),
			})
		}
		tipper = tx.GetTip().Tipper
	}

	return txv1beta1.Tip{
		Amount: parsedTipAmount,
		Tipper: tipper,
	}
}

func parseSignerInfo(signerData signing2.SignerData) []*txv1beta1.SignerInfo {
	parsedSignerInfo := []*txv1beta1.SignerInfo{}
	signerInfo := &txv1beta1.SignerInfo{
		PublicKey: signerData.PubKey,
		ModeInfo:  nil,
		Sequence:  signerData.Sequence,
	}
	parsedSignerInfo = append(parsedSignerInfo, signerInfo)
	return parsedSignerInfo
}

func parseTxMessages(tx authsigning.Tx) ([]*anypb.Any, error) {
	var parsedTxMsgs []*anypb.Any

	txPubKeys, err := tx.GetPubKeys()
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "Error on parsing TxData: ")
	}
	for _, txPubKey := range txPubKeys {
		parsedPubKey := anypb.Any{
			TypeUrl: sdk.MsgTypeURL(txPubKey),
			Value:   txPubKey.Bytes(),
		}
		parsedTxMsgs = append(parsedTxMsgs, &parsedPubKey)
	}
	return parsedTxMsgs, nil
}

func parseFeeAmount(tx authsigning.Tx) []*v1beta1.Coin {
	parsedFeeAmount := []*v1beta1.Coin{}
	for _, denom := range tx.GetFee().Denoms() {
		parsedFeeAmount = append(parsedFeeAmount, &v1beta1.Coin{
			Denom:  denom,
			Amount: tx.GetFee().AmountOf(denom).String(),
		})
	}
	return parsedFeeAmount
}

func parseAuthInfo(tx authsigning.Tx, signerData signing2.SignerData) *txv1beta1.AuthInfo {
	parsedTxTip := parseTxTip(tx)
	parsedFeeAmount := parseFeeAmount(tx)

	parsedSignerInfo := parseSignerInfo(signerData)

	return &txv1beta1.AuthInfo{
		SignerInfos: parsedSignerInfo,
		Fee: &txv1beta1.Fee{
			Amount:   parsedFeeAmount,
			GasLimit: tx.GetGas(),
			Payer:    string(tx.FeePayer()),
			Granter:  string(tx.FeeGranter()),
		},
		Tip: &parsedTxTip,
	}
}

func parseTxData(tx authsigning.Tx, signerData signing2.SignerData) (*signing2.TxData, error) {
	parsedTxMsgs, err := parseTxMessages(tx)
	if err != nil {
		return nil, err
	}

	txData := signing2.TxData{
		Body: &txv1beta1.TxBody{
			Messages:                    parsedTxMsgs,
			Memo:                        tx.GetMemo(),
			TimeoutHeight:               tx.GetTimeoutHeight(),
			ExtensionOptions:            nil,
			NonCriticalExtensionOptions: nil,
		},
		AuthInfo:                   parseAuthInfo(tx, signerData),
		BodyBytes:                  nil,
		AuthInfoBytes:              nil,
		BodyHasUnknownNonCriticals: false,
	}

	return &txData, err
}
