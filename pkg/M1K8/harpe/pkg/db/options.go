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
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

func (d *DB) CreateOption(oID, traderID, author string, channelType int, ticker, contractType, day, month, year string, price, starting, pt, poi, stop, underStart float32) (chan bool, string, bool, error) {

	chanMap.LoadOrStore(d.Guild, &sync.Map{})
	exists, exitChan := d.GetExitChanExists(fmt.Sprintf("%s_%s", oID, traderID))

	if exists {
		return exitChan, oID, exists, nil
	}

	if len(year) != 4 {
		return nil, "", false, errors.New("invalid Syntax - year is incorrect")
	}

	if len(month) > 2 || len(month) == 0 {
		return nil, "", false, errors.New("invalid Syntax - month is incorrect")
	}

	if len(day) > 2 || len(day) == 0 {
		return nil, "", false, errors.New("invalid Syntax - day is incorrect")
	}
	s := &Option{
		OptionAlertID:            fmt.Sprintf("%s_%s", oID, traderID),
		OptionGuildID:            d.Guild,
		TraderID:                 traderID,
		OptionTicker:             ticker,
		OptionUid:                oID,
		OptionDay:                day,
		OptionContractType:       contractType,
		OptionMonth:              month,
		OptionYear:               year,
		OptionStrike:             price,
		OptionStarting:           starting,
		ChannelType:              channelType,
		OptionCallTime:           time.Now(),
		OptionHighest:            starting,
		OptionUnderlyingPoI:      poi,
		OptionUnderlyingStop:     stop,
		OptionUnderlyingStarting: underStart,
		OptionUnderlyingPOIHit:   false,
		Caller:                   author,
	}

	_, err := d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(context.Background())

	if err != nil {
		log.Println(fmt.Sprintf("Unable to create option %v: %v.", oID, err.Error()))
		return nil, oID, false, err
	}

	return exitChan, oID, exists, nil
}

func (d *DB) RemoveOption(oID, traderID, contractType, day, month, year string, price float32) error {
	if len(year) != 4 {
		return errors.New("invalid Syntax - year is incorrect")
	}

	if len(month) > 2 || len(month) == 0 {
		return errors.New("invalid Syntax - month is incorrect")
	}

	if len(day) > 2 || len(day) == 0 {
		return errors.New("invalid Syntax - day is incorrect")
	}

	if len(month) == 1 {
		month = "0" + month
	}

	if len(day) == 1 {
		day = "0" + day
	}

	contxt := context.Background()

	s := &Option{
		OptionGuildID:  d.Guild,
		TraderID:       traderID,
		OptionUid:      oID,
		OptionStarting: 0,
		OptionCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("option_alert_id = ?", fmt.Sprintf("%s_%s", oID, traderID)).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to delete option %v: %v.", oID, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, fmt.Sprintf("%s_%s", oID, traderID))
	return nil
}

func (d *DB) RemoveOptionByCode(oID, traderID string) error {
	contxt := context.Background()

	s := &Option{
		OptionGuildID:  d.Guild,
		OptionUid:      oID,
		TraderID:       traderID,
		OptionStarting: 0,
		OptionCallTime: time.Time{},
	}
	_, err := d.db.NewDelete().Model(s).Where("option_alert_id = ?", fmt.Sprintf("%s_%s", oID, traderID)).Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to remove option %v: %v.", oID, err.Error()))
		return err
	}
	clearFromSyncMap(chanMap, d.Guild, fmt.Sprintf("%s_%s", oID, traderID))
	return nil
}

func (d *DB) GetOption(oID, traderID string) (*Option, error) {

	contxt := context.Background()

	s := &Option{
		OptionGuildID: d.Guild,
		TraderID:      traderID,
		OptionUid:     oID,
	}

	err := d.db.NewSelect().Model(s).Where("option_alert_id = ?", fmt.Sprintf("%s_%s", oID, traderID)).Scan(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to get option %v: %v.", oID, err.Error()))
		return nil, err
	}
	gMap, ok := chanMap.Load(d.Guild)
	if ok {
		gMapCast := gMap.(*sync.Map)
		_, ok := gMapCast.Load(fmt.Sprintf("%s_%s", oID, traderID))
		if !ok {
			log.Println(fmt.Sprintf("Unable to get alert channel for option %v. Please try recreating this alert, or calling !refresh then running this command again.", oID))
			return nil, err
		}
	}
	return s, nil
}

func (d *DB) OptionPOIHit(oID, traderID string) error {
	contxt := context.Background()
	s, err := d.GetOption(oID, traderID)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get option %v : %v", oID, err.Error()))
		return err
	}

	s.OptionUnderlyingPOIHit = true

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update option %v : %v", oID, err.Error()))
		return err
	}

	return nil
}

func (d *DB) OptionSetNewHigh(oID, traderID string, price float32) error {
	contxt := context.Background()

	s, err := d.GetOption(oID, traderID)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get Option %v : %v", oID, err.Error()))
		return err
	}

	s.OptionHighest = price

	_, err = d.db.NewInsert().Model(s).On("CONFLICT (option_alert_id) DO UPDATE").Exec(contxt)

	if err != nil {
		log.Println(fmt.Sprintf("Unable to update Option %v : %v", oID, err.Error()))
		return err
	}

	return nil
}

func (o Option) GetPctGain(highest float32) float32 {
	return ((highest - o.OptionStarting) / o.OptionStarting) * 100
}
