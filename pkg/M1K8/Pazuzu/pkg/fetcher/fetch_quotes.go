/*
 * Copyright 2022 M1K
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
package fetcher

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

var (
	finn string
)

func init() {
	finn = os.Getenv("finn")

	if finn == "" {
		panic("finnhub not defined!")
	}
}

func GetStock(stock, resolution string, from, to time.Time) (*finnhub.StockCandles, error) {
	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", finn)
	finnhubClient := finnhub.NewAPIClient(cfg).DefaultApi

	candles, req, err := finnhubClient.StockCandles(context.Background()).Symbol(strings.ToUpper(stock)).Resolution(resolution).From(from.Unix()).To(to.Unix()).Execute()

	if err != nil {
		if req != nil {
			log.Println(req.StatusCode)
			return nil, err
		} else {
			return nil, err
		}
	}

	if len(candles.GetC()) == 0 {
		return nil, errors.New("ticker not found")
	}

	return &candles, nil
	/*
		close
		high
		low
		open
		status
		timestamp
		volume
	*/
}

func GetCrypto(coin, resolution string, from, to time.Time) (*finnhub.CryptoCandles, error) {
	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", finn)
	finnhubClient := finnhub.NewAPIClient(cfg).DefaultApi

	candles, req, err := finnhubClient.CryptoCandles(context.Background()).Symbol("COINBASE:" + strings.ToUpper(coin) + "-USD").Resolution(resolution).From(from.Unix()).To(to.Unix()).Execute()

	if err != nil {
		if req != nil {
			log.Println(req.StatusCode)
			return nil, err
		} else {
			return nil, err
		}
	}

	if len(candles.GetC()) == 0 {
		return nil, errors.New("coin not found")
	}

	return &candles, nil
}
