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
package normalize

import (
	"github.com/Finnhub-Stock-API/finnhub-go/v2"
)

// this is dumb
type NormalCandleData struct {
	Date int64
	Data [4]float32
	//O    float32
	//C    float32
	//L    float32
	//H    float32

	/*/ List of open prices for returned candles.
	O *[]float32 `json:"o,omitempty"`
	// List of high prices for returned candles.
	H *[]float32 `json:"h,omitempty"`
	// List of low prices for returned candles.
	L *[]float32 `json:"l,omitempty"`
	// List of close prices for returned candles.
	C *[]float32 `json:"c,omitempty"`
	// List of volume data for returned candles.
	V *[]float32 `json:"v,omitempty"`
	// List of timestamp for returned candles.
	T *[]int64 `json:"t,omitempty"`
	// Status of the response. This field can either be ok or no_data.
	S *string `json:"s,omitempty"`
	*/
}

func NormalizeStocks(s *finnhub.StockCandles) ([]NormalCandleData, []float64, []int64) {
	// assume len() of each value is the same
	closing := make([]float64, 0, 0)
	times := make([]int64, 0, 0)
	res := make([]NormalCandleData, 0, 0)
	for i := range *s.O {
		time := *s.T
		o := *s.O
		h := *s.H
		l := *s.L
		c := *s.C
		res = append(res, NormalCandleData{
			Date: time[i],
			Data: [4]float32{o[i], c[i], l[i], h[i]},
		})
		closing = append(closing, float64(c[i]))
		times = append(times, time[i])
	}

	return res, closing, times
}

func NormalizeCrypto(c *finnhub.CryptoCandles) ([]NormalCandleData, []float64, []int64) {
	res := make([]NormalCandleData, 0, 0)
	closing := make([]float64, 0, 0)
	times := make([]int64, 0, 0)
	for i := range *c.O {
		time := *c.T
		o := *c.O
		h := *c.H
		l := *c.L
		c := *c.C
		res = append(res, NormalCandleData{
			Date: time[i],
			Data: [4]float32{o[i], c[i], l[i], h[i]},
		})
		closing = append(closing, float64(c[i]))
		times = append(times, time[i])
	}

	return res, closing, times
}
