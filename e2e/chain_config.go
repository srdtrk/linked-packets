package main

import (
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/srdtrk/linkedpackets/simapp/params"
)

var chainSpecs = []*interchaintest.ChainSpec{
	// -- WASMD --
	{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    "simd",
			ChainID: "simd-1",
			Images: []ibc.DockerImage{
				{
					Repository: "myimage", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "latest",        // FOR LOCAL IMAGE USE: Docker Image Tag
				},
			},
			Bin:           "simd",
			Bech32Prefix:  "cosmos",
			Denom:         "stake",
			GasPrices:     "0.00stake",
			GasAdjustment: 1.3,
			// cannot run wasmd commands without wasm encoding
			EncodingConfig:         params.MakeTestEncodingConfig(),
			TrustingPeriod:         "508h",
			NoHostMount:            false,
		},
	},	
	{
		ChainConfig: ibc.ChainConfig{
			Type:    "cosmos",
			Name:    "simd",
			ChainID: "simd-2",
			Images: []ibc.DockerImage{
				{
					Repository: "myimage", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "latest",        // FOR LOCAL IMAGE USE: Docker Image Tag
				},
			},
			Bin:           "simd",
			Bech32Prefix:  "cosmos",
			Denom:         "stake",
			GasPrices:     "0.00stake",
			GasAdjustment: 1.3,
			// cannot run wasmd commands without wasm encoding
			EncodingConfig:         params.MakeTestEncodingConfig(),
			TrustingPeriod:         "508h",
			NoHostMount:            false,
		},
	},
}
