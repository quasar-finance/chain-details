package internal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tendermint/tendermint/rpc/client/http"
)

type BondingID struct {
	BondID           string
	DepositorDetails []DepositorDetails
}

type DepositorDetails struct {
	Address          string
	BlockHeight      int64
	Amount           string
	PrimitiveAddress string
}

func ReplayChain(RPCAddress string, startingHeight, endHeight int64) ([]BondingID, error) {
	rpcClient, err := http.New(RPCAddress, "/websocket")
	if err != nil {
		return []BondingID{}, err
	}

	var bondingIDs []BondingID
	for i := startingHeight; i < endHeight; i++ {
		time.Sleep(time.Millisecond * 50)

		blockResults, err := rpcClient.BlockResults(context.Background(), &i)
		if err != nil {
			return bondingIDs, err
		}
		for _, j := range blockResults.TxsResults {
			var tempDepositorDetails []DepositorDetails
			var bondID string
			if strings.Contains(string(j.Data), "/cosmwasm.wasm.v1.MsgExecuteContract") {
				for o, k := range j.Events {
					if k.Type == "message" && string(k.Attributes[0].Value) == "/cosmwasm.wasm.v1.MsgExecuteContract" {
						if len(j.Events) >= o+3 {
							if j.Events[o+1].Type == "message" && string(j.Events[o+1].Attributes[0].Value) == "wasm" {
								if j.Events[o+2].Type == "coin_spent" {
									if j.Events[o+3].Type == "coin_received" && string(j.Events[o+3].Attributes[0].Value) == "quasar18a2u6az6dzw528rptepfg6n49ak6hdzkf8ewf0n5r0nwju7gtdgqamr7qu" {
										fmt.Println(string(j.Events[o+2].Attributes[0].Value), ":", string(j.Events[o+2].Attributes[1].Value))
										tempDepositorDetails = append(tempDepositorDetails, DepositorDetails{
											Address:          string(j.Events[o+2].Attributes[0].Value),
											Amount:           string(j.Events[o+2].Attributes[1].Value),
											BlockHeight:      i,
											PrimitiveAddress: "quasar18a2u6az6dzw528rptepfg6n49ak6hdzkf8ewf0n5r0nwju7gtdgqamr7qu",
										})
									}
								}
							}
						} else {
							fmt.Println("couldn't find the next 2 events on block height :", i)
						}
					}
				}

				for _, q := range j.Events {
					if q.Type == "wasm" && string(q.Attributes[0].Value) == "quasar18a2u6az6dzw528rptepfg6n49ak6hdzkf8ewf0n5r0nwju7gtdgqamr7qu" {
						bondID = string(q.Attributes[1].Value)
					}
				}

				if bondID != "" && len(tempDepositorDetails) > 0 {
					bondingIDs = append(bondingIDs, BondingID{
						BondID:           bondID,
						DepositorDetails: tempDepositorDetails,
					})
				}

			}

		}
	}
	return bondingIDs, nil
}
