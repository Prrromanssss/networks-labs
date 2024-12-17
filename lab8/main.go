package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type TransactionData struct {
	Hash     string `json:"hash"`
	From     string `json:"from"`
	To       string `json:"to,omitempty"`
	Value    string `json:"value"`
	Gas      uint64 `json:"gas"`
	GasPrice string `json:"gas_price"`
}

type BlockData struct {
	Number           uint64            `json:"number"`
	Time             uint64            `json:"time"`
	Difficulty       string            `json:"difficulty"`
	Hash             string            `json:"hash"`
	TransactionCount int               `json:"transaction_count"`
	Transactions     []TransactionData `json:"transactions,omitempty"`
}

func main() {
	infuraURL := "https://mainnet.infura.io/v3/133bce8536104f64a3944d1eba074047"

	firebaseURL := "https://my-test-project-e80ce-default-rtdb.europe-west1.firebasedatabase.app/blocks.json"

	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		log.Fatalln("Ethereum connection error:", err)
	}
	defer client.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastBlock uint64 = 0

	for range ticker.C {
		header, err := client.HeaderByNumber(context.Background(), nil)
		if err != nil {
			log.Println("Header by number error:", err)
			continue
		}

		latestBlockNumber := header.Number.Uint64()

		if latestBlockNumber > lastBlock {
			fmt.Printf("Find new block: %d\n", latestBlockNumber)

			block, err := client.BlockByNumber(context.Background(), header.Number)
			if err != nil {
				log.Println("Error getting block:", err)
				continue
			}

			chainID, err := client.NetworkID(context.Background())
			if err != nil {
				log.Println("Error getting ChainID:", err)
				continue
			}

			signer := types.LatestSignerForChainID(chainID)

			blockData := BlockData{
				Number:           block.Number().Uint64(),
				Time:             block.Time(),
				Difficulty:       block.Difficulty().String(),
				Hash:             block.Hash().Hex(),
				TransactionCount: len(block.Transactions()),
				Transactions:     []TransactionData{},
			}

			for _, tx := range block.Transactions() {
				from, err := types.Sender(signer, tx)
				if err != nil {
					log.Println("Error getting address:", err)
					continue
				}

				txData := TransactionData{
					Hash:     tx.Hash().Hex(),
					From:     from.Hex(),
					To:       "",
					Value:    tx.Value().String(),
					Gas:      tx.Gas(),
					GasPrice: tx.GasPrice().String(),
				}

				if tx.To() != nil {
					txData.To = tx.To().Hex()
				}

				blockData.Transactions = append(blockData.Transactions, txData)
			}

			jsonData, err := json.Marshal(blockData)
			if err != nil {
				log.Println("Error marshaling json:", err)
				continue
			}

			req, err := http.NewRequest("POST", firebaseURL, bytes.NewBuffer(jsonData))
			if err != nil {
				log.Println("Error creating new HTTP request:", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")

			clientHTTP := &http.Client{}
			resp, err := clientHTTP.Do(req)
			if err != nil {
				log.Println("Error sending data to Firebase:", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				fmt.Printf("Block data %d was successfully written to Firebase.\n", blockData.Number)
				lastBlock = latestBlockNumber
			} else {
				log.Printf("Cannot write data to Firebase. Status: %s\n", resp.Status)
			}
		} else {
			fmt.Println("New blocks weren't found.")
		}
	}

}
