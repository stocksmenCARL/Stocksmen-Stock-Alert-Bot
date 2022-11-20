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
package graph

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cinar/indicator"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/m1k8/kronos/pkg/M1K8/Pazuzu/pkg/fetcher"
	"github.com/m1k8/kronos/pkg/M1K8/Pazuzu/pkg/normalize"
	"github.com/m1k8/kronos/pkg/M1K8/Pazuzu/pkg/screenshot"
)

var loc, _ = time.LoadLocation("America/Detroit")

// TODO - WOrkaround
func getExtendedDeMarker(times []time.Time, max float32) *charts.Bar {
	bar := charts.NewBar()

	preY := make([]opts.BarData, 0)
	postY := make([]opts.BarData, 0)

	midnight := time.Now().In(loc).Format("2006/01/02")
	tim, _ := time.Parse("2006/01/02", midnight)

	bar = bar.SetXAxis(times)

	bar.SetGlobalOptions(
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        10000,
			XAxisIndex: []int{0},
		}),
	)

	if times[0].In(loc).After(tim.Add((9 * time.Hour) + (29 * time.Minute))) {
		// dont draw first line
	} else {
		preY = append(preY, opts.BarData{
			Value: []interface{}{tim.Format("3:04 PM"), 0},
		},
			opts.BarData{
				Name:  "a",
				Value: []interface{}{tim.Add(time.Minute).Format("3:04 PM"), 0},
			},
		)

		bar.AddSeries("pre", preY).SetSeriesOptions(
		//charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{
		//	Name:     "a",
		//	ValueDim: "maximum",
		//},
		//),
		)
	}

	if times[len(times)-1].In(loc).After(tim.Add(17 * time.Hour)) {
		postY = append(postY, opts.BarData{
			Name:  "b",
			Value: []interface{}{tim.Format("3:04 PM"), max},
		},
			opts.BarData{
				Name:  "b",
				Value: []interface{}{tim.Add(time.Minute).Format("3:04 PM"), max},
			},
		)
		bar.AddSeries("post", postY).SetSeriesOptions(
		//charts.WithMarkLineNameYAxisItemOpts(opts.MarkLineNameYAxisItem{
		//	Name:     "a",
		//	ValueDim: "maximum",
		//},
		//),
		)
	}

	return bar
}

func baseCandleGraph(candles []normalize.NormalCandleData, ticker string, timeframe string) *charts.Kline {
	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	var finalVal float32
	var initVal float32
	var maxVal float32

	if len(candles) <= 0 {
		return nil
	}
	initVal = candles[0].Data[1]
	times := []time.Time{}

	for i := 0; i < len(candles); i++ {
		tm := time.Unix(candles[i].Date, 0).In(loc)
		// add: volume, rsi - seperate Y axis
		// set scale to super low? opacity low? even worth it?

		switch timeframe {
		case "15":
			x = append(x, tm.Format(("3:04 PM")))
		case "60":
			fallthrough
		case "D":
			fallthrough
		case "W":
			x = append(x, tm.Format("01/02"))
		case "M":
			x = append(x, tm.Format("01/06"))
		}

		y = append(y, opts.KlineData{Value: candles[i].Data})
		times = append(times, tm)

		finalVal = candles[i].Data[1]

		if maxVal < finalVal {
			maxVal = finalVal
		}
	}

	pctChange := ((finalVal - initVal) / initVal) * 100
	var frame string
	switch timeframe {
	case "D":
		frame = "Daily"
	case "15":
		frame = "15 Min"
	case "60":
		frame = "Hourly"
	case "W":
		frame = "Weekly"
	case "M":
		frame = "Monthly"
	}

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: fmt.Sprintf("%v $%.2f (%.2f%%)\n%v", ticker, finalVal, pctChange, frame),
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        10000,
			XAxisIndex: []int{0},
		}),
	)

	kline.SetXAxis(x).AddSeries("Candles", y).SetSeriesOptions(
		charts.WithItemStyleOpts(opts.ItemStyle{
			Color:        "#008800",
			Color0:       "#880000",
			BorderColor:  "#008800",
			BorderColor0: "#880000",
		}),
	)

	if timeframe == "15" || timeframe == "60" {

		_ = getExtendedDeMarker(times, maxVal)
	}

	return kline
}

func bollingerBandsGraph(x []int64, m []float64, u []float64, l []float64) (ml, ul, ll *charts.Line) {
	ml = charts.NewLine()
	ul = charts.NewLine()
	ll = charts.NewLine()

	my := make([]opts.LineData, 0)
	uy := make([]opts.LineData, 0)
	ly := make([]opts.LineData, 0)

	for i := range m {
		my = append(my, opts.LineData{
			Value:      m[i],
			Symbol:     "none",
			YAxisIndex: 0,
		})
	}
	for i := range u {
		uy = append(uy, opts.LineData{
			Value:      u[i],
			Symbol:     "none",
			YAxisIndex: 0,
		})
	}
	for i := range l {
		ly = append(ly, opts.LineData{
			Value:      l[i],
			Symbol:     "none",
			YAxisIndex: 0,
		})
	}

	ml.SetXAxis(x).
		AddSeries("BM", my).SetSeriesOptions(
		charts.WithItemStyleOpts(
			opts.ItemStyle{
				Color: "#ffa000",
			},
		),
	)

	ul.SetXAxis(x).
		AddSeries("BU", uy).SetSeriesOptions(
		charts.WithItemStyleOpts(
			opts.ItemStyle{
				Color: "#ffb900",
			},
		),
		//charts.WithAreaStyleOpts(
		//	opts.AreaStyle{
		//		Color: "#ffb900",
		//	},
		//),
	) //

	ll.SetXAxis(x).
		AddSeries("BL", ly).SetSeriesOptions(
		charts.WithItemStyleOpts(
			opts.ItemStyle{
				Color: "#ff6100",
			},
		),
		//charts.WithAreaStyleOpts(
		//	opts.AreaStyle{
		//		Color:   "#ffffff",
		//		Opacity: 1,
		//	},
		//),
	)

	ll.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Ticker",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start:      0,
			End:        10000,
			XAxisIndex: []int{0},
		}),
	)

	return
}

func GetDStocksChart(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)
	stock, err := fetcher.GetStock(ticker, "D", time.Now().Add(-24*time.Hour*180), time.Now())

	if err != nil {
		return "", err
	}

	normalisedStock, c, t := normalize.NormalizeStocks(stock)

	w := baseCandleGraph(normalisedStock, ticker, "D")

	m, u, l := indicator.BollingerBands(c)
	ml, ll, ul := bollingerBandsGraph(t, m, u, l)
	w.Overlap(ll, ul, ml)

	page := components.NewPage()
	page.AddCharts(
		w,
	)

	f, err := os.Create("./" + ticker + ".html")
	if err != nil {
		return "", err

	}
	page.Render(io.MultiWriter(f))

	fileN := screenshot.TakeSS(ticker)

	return fileN, nil
}

func Get15MStocksChart(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)
	yesterday := time.Now().In(loc).Add(-time.Hour * 24).Format("2006/01/02")
	tim, _ := time.Parse("2006/01/02", yesterday)
	stock, err := fetcher.GetStock(ticker, "15", tim.Add(time.Hour*4), time.Now())

	if err != nil {
		return "", err
	}

	normalisedStock, c, t := normalize.NormalizeStocks(stock)

	w := baseCandleGraph(normalisedStock, ticker, "15")

	m, u, l := indicator.BollingerBands(c)
	ml, ll, ul := bollingerBandsGraph(t, m, u, l)
	w.Overlap(ll, ul, ml)

	page := components.NewPage()
	page.AddCharts(
		w,
	)

	f, err := os.Create("./" + ticker + ".html")
	if err != nil {
		return "", err

	}
	page.Render(io.MultiWriter(f))

	fileN := screenshot.TakeSS(ticker)

	return fileN, nil
}

func GetHStocksChart(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)
	yesterday := time.Now().In(loc).Add(-time.Hour * 336).Format("2006/01/02")
	tim, _ := time.Parse("2006/01/02", yesterday)
	stock, err := fetcher.GetStock(ticker, "60", tim.Add(time.Hour*4), time.Now())

	if err != nil {
		return "", err
	}

	normalisedStock, c, t := normalize.NormalizeStocks(stock)

	w := baseCandleGraph(normalisedStock, ticker, "60")

	m, u, l := indicator.BollingerBands(c)
	ml, ll, ul := bollingerBandsGraph(t, m, u, l)
	w.Overlap(ll, ul, ml)

	page := components.NewPage()
	page.AddCharts(
		w,
	)

	f, err := os.Create("./" + ticker + ".html")
	if err != nil {
		return "", err

	}
	page.Render(io.MultiWriter(f))

	fileN := screenshot.TakeSS(ticker)

	return fileN, nil
}

func GetDCryptoChart(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)
	crypto, err := fetcher.GetCrypto(ticker, "D", time.Now().Add(-24*time.Hour*180), time.Now())

	if err != nil {
		return "", err
	}

	normalisedCrypto, c, t := normalize.NormalizeCrypto(crypto)

	w := baseCandleGraph(normalisedCrypto, ticker, "D")

	m, u, l := indicator.BollingerBands(c)
	ml, ll, ul := bollingerBandsGraph(t, m, u, l)
	w.Overlap(ll, ul, ml)

	page := components.NewPage()
	page.AddCharts(
		w,
	)

	f, err := os.Create("./" + ticker + ".html")
	if err != nil {
		return "", err

	}
	page.Render(io.MultiWriter(f))

	fileN := screenshot.TakeSS(ticker)

	return fileN, nil
}

func Get15MCryptoChart(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)
	yesterday := time.Now().In(loc).Add(-time.Hour * 24).Format("2006/01/02")
	tim, _ := time.Parse("2006/01/02", yesterday)
	crypto, err := fetcher.GetCrypto(ticker, "15", tim, time.Now())

	if err != nil {
		return "", err
	}

	normalisedCrypto, c, t := normalize.NormalizeCrypto(crypto)

	w := baseCandleGraph(normalisedCrypto, ticker, "15")

	m, u, l := indicator.BollingerBands(c)
	ml, ll, ul := bollingerBandsGraph(t, m, u, l)
	w.Overlap(ll, ul, ml)

	page := components.NewPage()
	page.AddCharts(
		w,
	)

	f, err := os.Create("./" + ticker + ".html")
	if err != nil {
		return "", err

	}
	page.Render(io.MultiWriter(f))

	fileN := screenshot.TakeSS(ticker)

	return fileN, nil
}

func GetHCryptoChart(ticker string) (string, error) {
	ticker = strings.ToUpper(ticker)
	yesterday := time.Now().In(loc).Add(-time.Hour * 336).Format("2006/01/02")
	tim, _ := time.Parse("2006/01/02", yesterday)
	crypto, err := fetcher.GetCrypto(ticker, "60", tim, time.Now())

	if err != nil {
		return "", err
	}

	normalisedCrypto, c, t := normalize.NormalizeCrypto(crypto)

	w := baseCandleGraph(normalisedCrypto, ticker, "60")

	m, u, l := indicator.BollingerBands(c)
	ml, ll, ul := bollingerBandsGraph(t, m, u, l)
	w.Overlap(ll, ul, ml)

	page := components.NewPage()
	page.AddCharts(
		w,
	)

	f, err := os.Create("./" + ticker + ".html")
	if err != nil {
		return "", err

	}
	page.Render(io.MultiWriter(f))

	fileN := screenshot.TakeSS(ticker)

	return fileN, nil
}
