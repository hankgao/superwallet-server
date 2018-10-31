package mobile

import (
	"errors"
	"fmt"

	"github.com/hankgao/superwallet-server/server/mobile/bitcoin"
	"github.com/skycoin/skycoin-exchange/src/coin"
	"github.com/skycoin/skycoin-exchange/src/pp"
	"github.com/skycoin/skycoin-exchange/src/sknet"
	"github.com/skycoin/skycoin-exchange/src/wallet"
	"github.com/skycoin/skycoin/src/cipher"
)

type bitcoinCli struct {
	NodeAddr string
	fee      string // bitcoin fee
}

type btcSendParams struct {
	WalletID string
	ToAddr   string
	Amount   uint64
	Fee      uint64
}

func newBitcoin(nodeAddr string) *bitcoinCli {
	return &bitcoinCli{NodeAddr: nodeAddr, fee: "2000"} // default transaction fee is 2000
}

func (bn bitcoinCli) ValidateAddr(address string) error {
	_, err := cipher.BitcoinDecodeBase58Address(address)
	return err
}

func (bn bitcoinCli) GetBalance(addrs []string) (uint64, error) {
	return uint64(0), nil
}

func (bn bitcoinCli) CreateRawTx(txIns []coin.TxIn, getKey coin.GetPrivKey, txOuts interface{}) (string, error) {
	coin := bitcoin.Bitcoin{}
	rawtx, err := coin.CreateRawTx(txIns, txOuts)
	if err != nil {
		return "", fmt.Errorf("create raw tx failed:%v", err)
	}

	return coin.SignRawTx(rawtx, getKey)
}

func (bn bitcoinCli) BroadcastTx(rawtx string) (string, error) {
	// TODO: desn't have to ask superwallet server to proxy the operation
	return bitcoin.BroadcastTx(rawtx)
}

func (bn bitcoinCli) GetTransactionByID(txid string) (string, error) {
	return "", nil
}

// Send sends bitcoins to target address
func (bn bitcoinCli) Send(inputAddrs, privateKeys, targetAddress string, amount float64) (string, error) {
	return "", nil
}

func (bn bitcoinCli) PrepareTx(params interface{}) ([]coin.TxIn, interface{}, error) {
	p := params.(btcSendParams)

	addrs, err := wallet.GetAddresses(p.WalletID)
	if err != nil {
		return nil, nil, err
	}

	totalUtxos, err := bn.getOutputs(addrs)
	if err != nil {
		return nil, nil, err
	}

	utxos, bal, err := bn.getSufficientOutputs(totalUtxos, p.Amount+p.Fee)
	if err != nil {
		return nil, nil, err
	}

	txIns := make([]coin.TxIn, len(utxos))
	for i, u := range utxos {
		txIns[i] = coin.TxIn{
			Txid:    u.GetTxid(),
			Vout:    u.GetVout(),
			Address: u.GetAddress(),
		}
	}

	var txOut []bitcoin.TxOut
	chgAmt := bal - p.Amount - p.Fee
	chgAddr := addrs[0]
	if chgAmt > 0 {
		txOut = append(txOut,
			bn.makeTxOut(p.ToAddr, p.Amount),
			bn.makeTxOut(chgAddr, chgAmt))
	} else {
		txOut = append(txOut, bn.makeTxOut(p.ToAddr, p.Amount))
	}

	return txIns, txOut, nil
}

func (bn bitcoinCli) makeTxOut(addr string, value uint64) bitcoin.TxOut {
	return bitcoin.TxOut{
		Addr:  addr,
		Value: value,
	}
}

func (bn bitcoinCli) getSufficientOutputs(utxos []*pp.BtcUtxo, amt uint64) ([]*pp.BtcUtxo, uint64, error) {
	outMap := make(map[string][]*pp.BtcUtxo)
	for _, u := range utxos {
		outMap[u.GetAddress()] = append(outMap[u.GetAddress()], u)
	}

	allUtxos := []*pp.BtcUtxo{}
	var allBal uint64
	for _, utxos := range outMap {
		allBal += func(utxos []*pp.BtcUtxo) uint64 {
			var bal uint64
			for _, u := range utxos {
				if u.GetAmount() == 0 {
					continue
				}
				bal += u.GetAmount()
			}
			return bal
		}(utxos)

		allUtxos = append(allUtxos, utxos...)
		if allBal >= amt {
			return allUtxos, allBal, nil
		}
	}
	return nil, 0, errors.New("insufficient balance")
}

func (bn bitcoinCli) getOutputs(addrs []string) ([]*pp.BtcUtxo, error) {
	req := pp.GetUtxoReq{
		CoinType:  pp.PtrString("bitcoin"),
		Addresses: addrs,
	}
	res := pp.GetUtxoRes{}
	if err := sknet.EncryGet(bn.NodeAddr, "/get/utxos", req, &res); err != nil {
		return nil, err
	}

	if !res.Result.GetSuccess() {
		return nil, fmt.Errorf("get utxos failed: %v", res.Result.GetReason())
	}

	return res.BtcUtxos, nil
}
