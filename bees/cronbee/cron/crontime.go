/*
 *    Copyright (C) 2014      Stefan 'glaxx' Luecke
 *                  2014-2017 Christian Muehlhaeuser
 *
 *    This program is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU Affero General Public License as published
 *    by the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    This program is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU Affero General Public License for more details.
 *
 *    You should have received a copy of the GNU Affero General Public License
 *    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *    Authors:
 *		Stefan Luecke <glaxx@glaxx.net>
 *      Christian Muehlhaeuser <muesli@gmail.com>
 */

/*
 * TODO List:
 * - Test leap year behavior
 * - Test
 */

// Package cron allows you to schedule events.
package cron

import (
	"container/list"
	"time"

	log "github.com/sirupsen/logrus"
)

type crontime struct {
	second                []int
	minute                []int
	hour                  []int
	dow                   []int //Day of Week
	dom                   []int //Day of Month
	month                 []int
	calculatedTime        time.Time
	calculationInProgress bool
	eventList             list.List
}

// Returns the time.Duration until the next event.
func (c *crontime) DurationUntilNextEvent() time.Duration {
	return c.nextEvent().Sub(time.Now())
}

func (c *crontime) GetNextEvent() time.Time {
	return c.eventList.Front().Value.(time.Time)
}

func (c *crontime) nextEvent() time.Time {
	if !c.calculationInProgress && c.eventList.Len() == 0 {
		r := c.calculateEvent(time.Now())
		go c.fillList(r)
		return r
	} else if c.calculationInProgress && c.eventList.Len() == 0 {
		// shit just got real aka TODO
		panic("Shit")

	} else if c.eventList.Len() > 0 {
		e := c.eventList.Front()
		r := e.Value.(time.Time)
		c.eventList.Remove(e)
		go c.fillList(c.eventList.Back().Value.(time.Time))
		return r
	}
	panic("shit 2")
}

func (c *crontime) fillList(baseTime time.Time) {
	if c.eventList.Len() == 0 {
		c.eventList.PushBack(c.calculateEvent(baseTime))
	}
	for c.eventList.Len() < 5 {
		c.eventList.PushBack(c.calculateEvent(c.eventList.Back().Value.(time.Time)))
	}
}

func (c *crontime) setCalculationInProgress(set bool) {
	c.calculationInProgress = set
}

// This functions calculates the next event
func (c *crontime) calculateEvent(baseTime time.Time) time.Time {
	c.calculationInProgress = true
	defer c.setCalculationInProgress(false)
	baseTime = setNanoecond(baseTime, 10000)
	c.calculatedTime = baseTime // Ignore all Events in the Past & initial 'result'
	//c.calculatedTime = setNanoecond(c.calculatedTime, 10000)
	c.nextValidMonth(baseTime)
	c.nextValidDay(baseTime)
	c.nextValidHour(baseTime)
	c.nextValidMinute(baseTime)
	c.nextValidSecond(baseTime)
	log.Println("Cronbee has found a time stamp: ", c.calculatedTime)
	return c.calculatedTime
}

// Calculates the next valid Month based upon the previous results.
func (c *crontime) nextValidMonth(baseTime time.Time) {
	for _, mon := range c.month {
		if mon >= int(c.calculatedTime.Month()) {
			//log.Print("Inside Month", mon, c.calculatedTime)
			c.calculatedTime = setMonth(c.calculatedTime, mon)
			//log.Println(" :: and out", c.calculatedTime)
			return
		}
	}
	// If no result was found try it again in the following year
	c.calculatedTime = c.calculatedTime.AddDate(1, 0, 0)
	c.calculatedTime = setMonth(c.calculatedTime, c.month[0])
	//log.Println("Cronbee: Month", c.calculatedTime, baseTime, c.month)
	c.nextValidMonth(baseTime)
}

// Calculates the next valid Day based upon the previous results.
func (c *crontime) nextValidDay(baseTime time.Time) {
	for _, dom := range c.dom {
		if dom >= c.calculatedTime.Day() {
			for _, dow := range c.dow {
				if monthHasDow(dow, dom, int(c.calculatedTime.Month()), c.calculatedTime.Year()) {
					c.calculatedTime = setDay(c.calculatedTime, dom)
					//log.Println("Cronbee: Day-INS-1:", c.calculatedTime)
					return
				}
			}
		}
	} /* else {
		for _, dow := range c.dow {
			if monthHasDow(dow, dom, int(c.calculatedTime.Month()), c.calculatedTime.Year()){
				c.calculatedTime = setDay(c.calculatedTime, dom)
				log.Println("Cronbee: Day-INS-2:", c.calculatedTime)
				return
			}
		}
	}*/
	// If no result was found try it again in the following month.
	c.calculatedTime = c.calculatedTime.AddDate(0, 1, 0)
	c.calculatedTime = setDay(c.calculatedTime, c.dom[0])
	//log.Println("Cronbee: Day", c.calculatedTime, baseTime)
	c.nextValidMonth(baseTime)
	c.nextValidDay(baseTime)
}

// Calculates the next valid Hour based upon the previous results.
func (c *crontime) nextValidHour(baseTime time.Time) {
	for _, hour := range c.hour {
		if c.calculatedTime.Day() == baseTime.Day() {
			if !hasPassed(hour, c.calculatedTime.Hour()) {
				c.calculatedTime = setHour(c.calculatedTime, hour)
				return
			}
		} else {
			c.calculatedTime = setHour(c.calculatedTime, hour)
			return
		}
	}
	// If no result was found try it again in the following day.
	c.calculatedTime = c.calculatedTime.AddDate(0, 0, 1)    // <-|
	c.calculatedTime = setHour(c.calculatedTime, c.hour[0]) //   |
	//log.Println("Cronbee: Hour", c.calculatedTime, baseTime) // |
	c.nextValidMonth(baseTime) // May trigger a new month --------|
	c.nextValidDay(baseTime)
	c.nextValidHour(baseTime)
}

// Calculates the next valid Minute based upon the previous results.
func (c *crontime) nextValidMinute(baseTime time.Time) {
	for _, min := range c.minute {
		if c.calculatedTime.Hour() == baseTime.Hour() {
			if !hasPassed(min, c.calculatedTime.Minute()) {
				c.calculatedTime = setMinute(c.calculatedTime, min)
				return
			}
		} else {
			c.calculatedTime = setMinute(c.calculatedTime, min)
			return
		}
	}
	c.calculatedTime = c.calculatedTime.Add(1 * time.Hour)
	c.calculatedTime = setMinute(c.calculatedTime, c.minute[0])
	//log.Println("Cronbee: Minute", c.calculatedTime, baseTime)
	c.nextValidHour(baseTime)
	c.nextValidMinute(baseTime)
}

// Calculates the next valid Second based upon the previous results.
func (c *crontime) nextValidSecond(baseTime time.Time) {
	for _, sec := range c.second {
		if !c.minuteHasPassed(baseTime) {
			// check if sec is in the past. <= prevents triggering the same event twice
			if sec > c.calculatedTime.Second() {
				c.calculatedTime = setSecond(c.calculatedTime, sec)
				return
			}
		} else {
			c.calculatedTime = setSecond(c.calculatedTime, sec)
			return
		}
	}
	c.calculatedTime = c.calculatedTime.Add(1 * time.Minute)
	c.calculatedTime = setSecond(c.calculatedTime, 0)
	//log.Println("Cronbee: Second", c.calculatedTime, baseTime)
	c.nextValidMinute(baseTime)
	c.nextValidSecond(baseTime)
}

func (c *crontime) minuteHasPassed(baseTime time.Time) bool {
	if c.calculatedTime.Year() > baseTime.Year() {
		return true
	} else if c.calculatedTime.Month() > baseTime.Month() {
		return true
	} else if c.calculatedTime.Day() > baseTime.Day() {
		return true
	} else if c.calculatedTime.Hour() > baseTime.Hour() {
		return true
	} else if c.calculatedTime.Minute() > baseTime.Minute() {
		return true
	}
	return false
}

func hasPassed(value, tstamp int) bool {
	return value < tstamp
}

func monthHasDom(dom, month, year int) bool {
	switch month {
	case 1:
		return dom <= 31
	case 2:
		if isLeapYear(year) {
			return dom <= 29
		}
		return dom <= 28
	case 3:
		return dom <= 31
	case 4:
		return dom <= 30
	case 5:
		return dom <= 31
	case 6:
		return dom <= 30
	case 7:
		return dom <= 31
	case 8:
		return dom <= 31
	case 9:
		return dom <= 30
	case 10:
		return dom <= 31
	case 11:
		return dom <= 30
	case 12:
		return dom <= 31
	default:
		panic("strange thingys are happening!")
	}
}

// Check if the combination of day(of month), month and year is the weekday dow.
func monthHasDow(dow, dom, month, year int) bool {
	if !monthHasDom(dom, month, year) {
		return false
	}
	Nday := dom % 7
	var Nmonth int
	switch month {
	case 1:
		Nmonth = 0
	case 2:
		Nmonth = 3
	case 3:
		Nmonth = 3
	case 4:
		Nmonth = 6
	case 5:
		Nmonth = 1
	case 6:
		Nmonth = 4
	case 7:
		Nmonth = 6
	case 8:
		Nmonth = 2
	case 9:
		Nmonth = 5
	case 10:
		Nmonth = 0
	case 11:
		Nmonth = 3
	case 12:
		Nmonth = 5
	}
	var Nyear int
	temp := year % 100
	if temp != 0 {
		Nyear = (temp + (temp / 4)) % 7
	} else {
		Nyear = 0
	}
	Ncent := (3 - ((year / 100) % 4)) * 2
	var Nsj int
	if isLeapYear(year) {
		Nsj = -1
	} else {
		Nsj = 0
	}
	W := (Nday + Nmonth + Nyear + Ncent + Nsj) % 7
	return dow == W
}

func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

//
func setMonth(tstamp time.Time, month int) time.Time {
	if month > 12 || month < 1 {
		panic("ERROR Month")
	}
	return tstamp.AddDate(0, -absolute(int(tstamp.Month()), month), 0)
}

func setDay(tstamp time.Time, day int) time.Time {
	if day > 31 || day < 1 {
		panic("ERROR Day")
	}
	return tstamp.AddDate(0, 0, -absolute(tstamp.Day(), day))
}

func setHour(tstamp time.Time, hour int) time.Time {
	if hour >= 24 || hour < 0 {
		panic("ERROR Hour")
	}
	return tstamp.Add(time.Duration(-absolute(tstamp.Hour(), hour)) * time.Hour)
}

func setMinute(tstamp time.Time, minute int) time.Time {
	if minute >= 60 || minute < 0 {
		panic("ERROR Minute")
	}
	return tstamp.Add(time.Duration(-absolute(tstamp.Minute(), minute)) * time.Minute)
}

func setSecond(tstamp time.Time, second int) time.Time {
	if second >= 60 || second < 0 {
		panic("ERROR Second")
	}
	return tstamp.Add(time.Duration(-absolute(tstamp.Second(), second)) * time.Second)
}

func setNanoecond(tstamp time.Time, nanosecond int) time.Time {
	return tstamp.Add(time.Duration(-absolute(tstamp.Nanosecond(), nanosecond)) * time.Nanosecond)
}

func absolute(a, b int) int {
	return a - b
}
