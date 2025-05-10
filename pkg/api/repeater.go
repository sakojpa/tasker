package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Daily RuleType = iota + 1
	Weekly
	Monthly
	Yearly
)

type RuleType int

// TaskRepeat defines task repetition rule type and next scheduled date.
type TaskRepeat struct {
	Type     RuleType
	Repeat   interface{}
	NextDate string
}

func validateRepeatRule(now time.Time, dstart string, repeat string) (taskRepeat *TaskRepeat, err error) {
	if repeat == "" {
		return nil, fmt.Errorf("interval should be specified: repeat='%s'", repeat)
	}
	date, err := time.Parse(dateFormat, dstart)
	if err != nil {
		return nil, fmt.Errorf("date format error: '%v'", err)
	}
	taskInfo := &TaskRepeat{}

	switch {
	case strings.HasPrefix(repeat, "y"):
		taskInfo.Type = Yearly
		for {
			date = date.AddDate(1, 0, 0)
			if afterNow(date, now) {
				taskInfo.NextDate = date.Format(dateFormat)
				return taskInfo, nil
			}
		}
	case strings.HasPrefix(repeat, "d"):
		taskInfo.Type = Daily
		repeatStr := strings.TrimSpace(strings.TrimPrefix(repeat, "d"))
		repeatInt, err := strconv.Atoi(repeatStr)
		if err != nil || repeatInt < 1 || repeatInt > 400 {
			return nil, fmt.Errorf("incorrect days interval: '%s'", repeatStr)
		}
		taskInfo.Repeat = repeatInt

		for {
			date = date.AddDate(0, 0, taskInfo.Repeat.(int))
			if afterNow(date, now) {
				taskInfo.NextDate = date.Format(dateFormat)
				return taskInfo, nil
			}
		}

	case strings.HasPrefix(repeat, "w"):
		taskInfo.Type = Weekly
		repeatDofW := strings.TrimSpace(strings.TrimPrefix(repeat, "w "))
		weekdays := strings.Replace(repeatDofW, "7", "0", -1)
		for _, wday := range strings.Split(weekdays, ",") {
			i, err := strconv.Atoi(wday)
			if err != nil || i < 0 || i > 6 {
				return nil, fmt.Errorf("weekdays '%s' is not a weekday", wday)
			}
		}
		for {
			date = date.AddDate(0, 0, 1)
			if afterNow(date, now) && strings.Contains(weekdays, strconv.Itoa(int(date.Weekday()))) {
				taskInfo.NextDate = date.Format(dateFormat)
				return taskInfo, nil
			}
		}

	case strings.HasPrefix(repeat, "m"):
		taskInfo.Type = Monthly
		repeatDofMStr := strings.TrimSpace(strings.TrimPrefix(repeat, "m "))
		parts := strings.Split(repeatDofMStr, " ")
		switch len(parts) {
		case 1, 2:
			var dayMap [32]bool
			var monthMap [13]bool
			var ultimateDay bool
			var penultimateDay bool
			daySlice := strings.Split(parts[0], ",")
			for _, dayStr := range daySlice {
				dayInt, err := strconv.Atoi(dayStr)
				if err != nil || dayInt < -2 || dayInt > 31 || dayInt == 0 {
					return nil, fmt.Errorf("invalid day of month: '%v'", err)
				}
				if dayInt > 0 {
					dayMap[dayInt] = true
				} else if dayInt == -1 {
					ultimateDay = true
				} else {
					penultimateDay = true
				}
			}
			if len(parts) == 2 {
				monthSlice := strings.Split(parts[1], ",")
				for _, monthStr := range monthSlice {
					monthInt, err := strconv.Atoi(monthStr)
					if err != nil || monthInt < 1 || monthInt > 12 {
						return nil, fmt.Errorf("invalid month: '%v'", err)
					}
					monthMap[monthInt] = true
				}
			} else {
				for monthInt := 1; monthInt <= 12; monthInt++ {
					monthMap[monthInt] = true
				}
			}

			for {
				date = date.AddDate(0, 0, 1)
				if afterNow(date, now) && monthMap[date.Month()] {
					if dayMap[date.Day()] {
						taskInfo.NextDate = date.Format(dateFormat)
						return taskInfo, nil
					}
					if ultimateDay {
						lastDay := getLastDayOfMonth(date.Month(), date.Year())
						if checkDaysEquality(date, lastDay, dateFormat) {
							taskInfo.NextDate = date.Format(dateFormat)
							return taskInfo, nil
						}
					}
					if penultimateDay {
						lastDay := getLastDayOfMonth(date.Month(), date.Year())
						beforeLastDay := lastDay.AddDate(0, 0, -1)
						if checkDaysEquality(date, beforeLastDay, dateFormat) {
							taskInfo.NextDate = date.Format(dateFormat)
							return taskInfo, nil
						}
					}
				}
			}
		default:
			return nil, fmt.Errorf("invalid month interval: '%s'", repeat)
		}
	default:
		return nil, fmt.Errorf("invalid repeat rule")
	}
}
