// Copyright 2024 BSC
// ABI definitions

package txfilter

const buyerABIJSON = `[
	{
		"type": "function",
		"name": "buyWithBNBAndBribeTo",
		"stateMutability": "payable",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "minQuoteAmount", "type": "uint256"},
			{"name": "minTokenAmount", "type": "uint256"},
			{"name": "bribeTo", "type": "address"},
			{"name": "bribeAmount", "type": "uint256"}
		],
		"outputs": []
	}
]`

const erc20ABIJSON = `[
	{
		"constant": true,
		"inputs": [{"name": "_owner", "type": "address"}],
		"name": "balanceOf",
		"outputs": [{"name": "balance", "type": "uint256"}],
		"type": "function"
	},
	{
		"constant": false,
		"inputs": [
			{"name": "_spender", "type": "address"},
			{"name": "_value", "type": "uint256"}
		],
		"name": "approve",
		"outputs": [{"name": "", "type": "bool"}],
		"type": "function"
	}
]`

const fourmemeABIJSON = `[
	{
		"type": "function",
		"name": "sellToken",
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "amount", "type": "uint256"},
			{"name": "minBNBOut", "type": "uint256"}
		],
		"outputs": [],
		"stateMutability": "nonpayable"
	}
]`
