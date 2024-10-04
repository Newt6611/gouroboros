// Copyright 2024 Blink Labs Software
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package shelley

import (
	"fmt"
	"math/big"

	"github.com/blinklabs-io/gouroboros/cbor"
)

type ShelleyProtocolParameters struct {
	cbor.StructAsArray
	MinFeeA            uint
	MinFeeB            uint
	MaxBlockBodySize   uint
	MaxTxSize          uint
	MaxBlockHeaderSize uint
	KeyDeposit         uint
	PoolDeposit        uint
	MaxEpoch           uint
	NOpt               uint
	A0                 *cbor.Rat
	Rho                *cbor.Rat
	Tau                *cbor.Rat
	Decentralization   *cbor.Rat
	Nonce              *Nonce
	ProtocolMajor      uint
	ProtocolMinor      uint
	MinUtxoValue       uint
}

func (p *ShelleyProtocolParameters) Update(paramUpdate *ShelleyProtocolParameterUpdate) {
	if paramUpdate.MinFeeA != nil {
		p.MinFeeA = *paramUpdate.MinFeeA
	}
	if paramUpdate.MinFeeB != nil {
		p.MinFeeB = *paramUpdate.MinFeeB
	}
	if paramUpdate.MaxBlockBodySize != nil {
		p.MaxBlockBodySize = *paramUpdate.MaxBlockBodySize
	}
	if paramUpdate.MaxTxSize != nil {
		p.MaxTxSize = *paramUpdate.MaxTxSize
	}
	if paramUpdate.MaxBlockHeaderSize != nil {
		p.MaxBlockHeaderSize = *paramUpdate.MaxBlockHeaderSize
	}
	if paramUpdate.KeyDeposit != nil {
		p.KeyDeposit = *paramUpdate.KeyDeposit
	}
	if paramUpdate.PoolDeposit != nil {
		p.PoolDeposit = *paramUpdate.PoolDeposit
	}
	if paramUpdate.MaxEpoch != nil {
		p.MaxEpoch = *paramUpdate.MaxEpoch
	}
	if paramUpdate.NOpt != nil {
		p.NOpt = *paramUpdate.NOpt
	}
	if paramUpdate.A0 != nil {
		p.A0 = paramUpdate.A0
	}
	if paramUpdate.Rho != nil {
		p.Rho = paramUpdate.Rho
	}
	if paramUpdate.Tau != nil {
		p.Tau = paramUpdate.Tau
	}
	if paramUpdate.Decentralization != nil {
		p.Decentralization = paramUpdate.Decentralization
	}
	if paramUpdate.ProtocolVersion != nil {
		p.ProtocolMajor = paramUpdate.ProtocolVersion.Major
		p.ProtocolMinor = paramUpdate.ProtocolVersion.Minor
	}
	if paramUpdate.Nonce != nil {
		p.Nonce = paramUpdate.Nonce
	}
	if paramUpdate.MinUtxoValue != nil {
		p.MinUtxoValue = *paramUpdate.MinUtxoValue
	}
}

func (p *ShelleyProtocolParameters) UpdateFromGenesis(genesis *ShelleyGenesis) {
	genesisParams := genesis.ProtocolParameters
	p.MinFeeA = genesisParams.MinFeeA
	p.MinFeeB = genesisParams.MinFeeB
	p.MaxBlockBodySize = genesisParams.MaxBlockBodySize
	p.MaxTxSize = genesisParams.MaxTxSize
	p.MaxBlockHeaderSize = genesisParams.MaxBlockHeaderSize
	p.KeyDeposit = genesisParams.KeyDeposit
	p.PoolDeposit = genesisParams.PoolDeposit
	p.MaxEpoch = genesisParams.MaxEpoch
	p.NOpt = genesisParams.NOpt
	if genesisParams.A0 != nil {
		p.A0 = &cbor.Rat{Rat: new(big.Rat).Set(genesisParams.A0.Rat)}
	}
	if genesisParams.Rho != nil {
		p.Rho = &cbor.Rat{Rat: new(big.Rat).Set(genesisParams.Rho.Rat)}
	}
	if genesisParams.Tau != nil {
		p.Tau = &cbor.Rat{Rat: new(big.Rat).Set(genesisParams.Tau.Rat)}
	}
	if genesisParams.Decentralization != nil {
		p.Decentralization = &cbor.Rat{Rat: new(big.Rat).Set(genesisParams.Decentralization.Rat)}
	}
	p.ProtocolMajor = genesisParams.ProtocolVersion.Major
	p.ProtocolMinor = genesisParams.ProtocolVersion.Minor
	p.MinUtxoValue = genesisParams.MinUtxoValue
	// TODO:
	//p.Nonce              *cbor.Rat
}

type ShelleyProtocolParametersProtocolVersion struct {
	cbor.StructAsArray
	Major uint
	Minor uint
}

type ShelleyProtocolParameterUpdate struct {
	cbor.DecodeStoreCbor
	MinFeeA            *uint                                     `cbor:"0,keyasint"`
	MinFeeB            *uint                                     `cbor:"1,keyasint"`
	MaxBlockBodySize   *uint                                     `cbor:"2,keyasint"`
	MaxTxSize          *uint                                     `cbor:"3,keyasint"`
	MaxBlockHeaderSize *uint                                     `cbor:"4,keyasint"`
	KeyDeposit         *uint                                     `cbor:"5,keyasint"`
	PoolDeposit        *uint                                     `cbor:"6,keyasint"`
	MaxEpoch           *uint                                     `cbor:"7,keyasint"`
	NOpt               *uint                                     `cbor:"8,keyasint"`
	A0                 *cbor.Rat                                 `cbor:"9,keyasint"`
	Rho                *cbor.Rat                                 `cbor:"10,keyasint"`
	Tau                *cbor.Rat                                 `cbor:"11,keyasint"`
	Decentralization   *cbor.Rat                                 `cbor:"12,keyasint"`
	Nonce              *Nonce                                    `cbor:"13,keyasint"`
	ProtocolVersion    *ShelleyProtocolParametersProtocolVersion `cbor:"14,keyasint"`
	MinUtxoValue       *uint                                     `cbor:"15,keyasint"`
}

func (ShelleyProtocolParameterUpdate) IsProtocolParameterUpdate() {}

func (u *ShelleyProtocolParameterUpdate) UnmarshalCBOR(data []byte) error {
	return u.UnmarshalCbor(data, u)
}

const (
	NonceType0 = 0
	NonceType1 = 1
)

var NeutralNonce = Nonce{
	Type: NonceType0,
}

type Nonce struct {
	cbor.StructAsArray
	Type  uint
	Value [32]byte
}

func (n *Nonce) UnmarshalCBOR(data []byte) error {
	nonceType, err := cbor.DecodeIdFromList(data)
	if err != nil {
		return err
	}

	n.Type = uint(nonceType)

	switch nonceType {
	case NonceType0:
		// Value uses default value
	case NonceType1:
		if err := cbor.DecodeGeneric(data, n); err != nil {
			fmt.Printf("Nonce decode error: %+v\n", data)
			return err
		}
	default:
		return fmt.Errorf("unsupported nonce type %d", nonceType)
	}
	return nil
}
