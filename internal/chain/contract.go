package chain

import (
	"boralabs/config"
	boraLabsErr "boralabs/pkg/error"
	"boralabs/pkg/util"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strconv"
	"time"
)

const (
	ContractNameDao      = "dao"
	ContractNameGovernor = "governor"
	retryCnt             = 3
	FuncPastTotalSupply  = "getPastTotalSupply"
)

type Contract struct {
	name    string
	address string
	ecl     *ethclient.Client
	bound   *bind.BoundContract
	events  map[string]abi.Event
}

var (
	GovCont         *Contract
	DaoCont         *Contract
	daoAddress      = ""
	governorAddress = ""
)

func init() {
	var err error
	// validate contract start block
	fromBlock := config.C.GetString("fromBlock")
	if _, err = strconv.ParseInt(fromBlock, 10, 64); err != nil {
		panic(fmt.Sprintf(boraLabsErr.WrongFromBlockNumber, fromBlock))
	}

	daoAddress = config.C.GetString("daoAddress")
	governorAddress = config.C.GetString("governorAddress")
	checkValues := []string{daoAddress, governorAddress}
	for _, value := range checkValues {
		emptyCheck(value)
	}

	GovCont, err = NewContract(ContractNameGovernor)
	if err != nil {
		panic(err)
	}
	DaoCont, err = NewContract(ContractNameDao)
	if err != nil {
		panic(err)
	}
}

func emptyCheck(bindAddress string) {
	if bindAddress == "" {
		panic(boraLabsErr.EmptyConfigValue)
	}
}

func NewContract(name string) (*Contract, error) {
	var filePath, address string
	switch name {
	case ContractNameDao:
		filePath, address = "abi/BoraLabsDaoToken.json", daoAddress
	case ContractNameGovernor:
		filePath, address = "abi/BoraLabsGovernor.json", governorAddress
	default:
		return nil, errors.New(boraLabsErr.FailedNewContract)
	}

	var a abi.ABI
	if err := util.ReadJSONFile(&a, filePath); err != nil {
		return nil, err
	}

	ecl, err := ethclient.DialContext(context.Background(), config.C.GetString("rpcEndpoint"))
	if err != nil {
		panic(errors.New(fmt.Sprintf("Dialing error %v", err)))
	}

	return &Contract{
		name:    name,
		address: address,
		ecl:     ecl,
		bound:   bind.NewBoundContract(common.HexToAddress(address), a, ecl, ecl, ecl),
		events:  a.Events,
	}, nil
}

func GetPastTotalSupply(voteStart time.Time) (result []interface{}, error error) {
	error = DaoCont.bound.Call(nil, &result, FuncPastTotalSupply, big.NewInt(voteStart.Unix()))
	return
}

func (c *Contract) FilterLogs(eventSignature string, fromBlock uint64) ([]types.Log, error) {
	return c.ecl.FilterLogs(context.Background(), filterQuery(eventSignature, c.address, fromBlock))
}

func (c *Contract) SubscribeFilterLogs(eventSignature string, fromBlock uint64, logsCh chan types.Log) (ethereum.Subscription, error) {
	return c.ecl.SubscribeFilterLogs(context.Background(), filterQuery(eventSignature, c.address, fromBlock), logsCh)
}

func (c *Contract) BlockByNumber(blockNumber *big.Int) (*types.Block, error) {
	var bl *types.Block
	var err error
	for i := 0; i < retryCnt; i++ {
		bl, err = c.ecl.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			continue
		} else {
			return bl, err
		}
	}
	return bl, err
}

func (c *Contract) UnpackLogData(out any, evtName string, data types.Log) error {
	err := c.bound.UnpackLog(out, evtName, data)
	if err != nil {
		return err
	}
	return nil
}

func filterQuery(eventSignature string, address string, block uint64) ethereum.FilterQuery {
	return ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(block)),
		ToBlock:   nil,
		Addresses: []common.Address{common.HexToAddress(address)},
		Topics: [][]common.Hash{
			{util.GenerateEventHash(eventSignature)},
		},
	}
}
