// Copyright 2019 ChainSafe Systems (ON) Corp.
// This file is part of gossamer.
//
// The gossamer library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The gossamer library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the gossamer library. If not, see <http://www.gnu.org/licenses/>.

package dev

import (
	"github.com/ChainSafe/gossamer/lib/runtime/wasmer"
	log "github.com/ChainSafe/log15"
)

var (
	// GlobalConfig

	// DefaultName is the node name
	DefaultName = string("Gossamer")
	// DefaultID is the chain ID
	DefaultID = string("dev")
	// DefaultConfig is the toml configuration path
	DefaultConfig = string("./chain/dev/config.toml")
	// DefaultBasePath is the node base directory path
	DefaultBasePath = string("~/.gossamer/dev")

	// DefaultMetricsPort is the metrics server port
	DefaultMetricsPort = uint32(9876)

	// DefaultLvl is the default log level
	DefaultLvl = log.LvlInfo

	// InitConfig

	// DefaultGenesis is the default genesis configuration path
	DefaultGenesis = string("./chain/dev/genesis-spec.json")

	// AccountConfig

	// DefaultKey is the default account key
	DefaultKey = string("alice")
	// DefaultUnlock is the account to unlock
	DefaultUnlock = string("")

	// CoreConfig

	// DefaultAuthority is true if the node is a block producer and a grandpa authority
	DefaultAuthority = true
	// DefaultRoles Default node roles
	DefaultRoles = byte(4) // authority node (see Table D.2)
	// DefaultBabeAuthority is true if the node is a block producer (overwrites previous settings)
	DefaultBabeAuthority = true
	// DefaultGrandpaAuthority is true if the node is a grandpa authority (overwrites previous settings)
	DefaultGrandpaAuthority = true
	// DefaultWasmInterpreter is the name of the wasm interpreter to use by default
	DefaultWasmInterpreter = wasmer.Name

	// NetworkConfig

	// DefaultNetworkPort network port
	DefaultNetworkPort = uint32(7001)
	// DefaultNetworkBootnodes network bootnodes
	DefaultNetworkBootnodes = []string(nil)
	// DefaultNoBootstrap disables bootstrap
	DefaultNoBootstrap = false
	// DefaultNoMDNS disables mDNS discovery
	DefaultNoMDNS = false

	// RPCConfig

	// DefaultRPCHTTPHost rpc host
	DefaultRPCHTTPHost = string("localhost")
	// DefaultRPCHTTPPort rpc port
	DefaultRPCHTTPPort = uint32(8545)
	// DefaultRPCModules rpc modules
	DefaultRPCModules = []string{"system", "author", "chain", "state", "rpc", "grandpa"}
	// DefaultRPCWSPort rpc websocket port
	DefaultRPCWSPort = uint32(8546)
	// DefaultRPCEnabled enables the RPC server
	DefaultRPCEnabled = true
	// DefaultWSEnabled enables the WS server
	DefaultWSEnabled = true
)
