/*
 * Copyright 2021 M1K
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package db

import (
	"time"

	"github.com/uptrace/bun"
)

type Alert interface {
	GetPctGain(closing float32) float32
}

type Stock struct {
	StockAlertID  string `bun:",pk"`
	StockGuildID  string
	TraderID      string
	StockTicker   string
	StockStarting float32
	StockEPt      float32
	StockSPt      float32
	StockExpiry   int64
	StockHighest  float32
	StockPoI      float32
	StockStop     float32
	ChannelType   int
	Caller        string
	StockPOIHit   bool
	StockCallTime time.Time
}

type Short struct {
	ShortAlertID  string `bun:",pk"`
	ShortGuildID  string
	TraderID      string
	ShortTicker   string
	ShortStarting float32
	ShortSPt      float32
	ShortEPt      float32
	ShortExpiry   int64
	ShortLowest   float32
	ShortPoI      float32
	ShortStop     float32
	ChannelType   int
	Caller        string
	ShortPOIHit   bool
	ShortCallTime time.Time
}

type Option struct {
	OptionAlertID            string `bun:",pk"`
	OptionGuildID            string
	TraderID                 string
	OptionTicker             string
	OptionUid                string
	OptionContractType       string
	OptionDay                string
	OptionMonth              string
	OptionYear               string
	OptionStrike             float32
	OptionStarting           float32
	OptionHighest            float32
	OptionUnderlyingPoI      float32
	OptionUnderlyingStop     float32
	OptionUnderlyingStarting float32
	ChannelType              int
	Caller                   string
	OptionUnderlyingPOIHit   bool
	OptionCallTime           time.Time
}

type Crypto struct {
	CryptoAlertID  string `bun:",pk"`
	CryptoGuildID  string
	TraderID       string
	CryptoCoin     string
	CryptoStarting float32
	CryptoSPt      float32
	CryptoEPt      float32
	CryptoExpiry   int64
	CryptoHighest  float32
	CryptoStop     float32
	CryptoPoI      float32
	ChannelType    int
	Caller         string
	CryptoPOIHit   bool
	CryptoCallTime time.Time
}

type DB struct {
	Guild string
	db    *bun.DB
}
