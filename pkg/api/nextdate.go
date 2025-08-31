package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

// NextDate обработка правил вычисления следующей даты.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("empty repeat rule")
	}

	date, err := time.Parse(DateFormat, dstart)

	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

	parts := strings.Fields(repeat)

	if len(parts) == 0 {
		return "", fmt.Errorf("invalid repeat format")
	}

	ruleType := parts[0]
	switch ruleType {
	case "d":
		return handleDailyRule(now, date, parts)
	case "y":
		return handleYearlyRule(now, date)
	case "w":
		return handleWeeklyRule(now, date, parts)
	case "m":
		return handleMonthlyRule(now, date, parts)
	default:
		return "", fmt.Errorf("unsupported repeat format %s", ruleType)
	}
}

// handleMonthlyRule обработка месячных правил для вычисления следующей даты
func handleMonthlyRule(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", fmt.Errorf("missing days for monthly rule")
	}

	daysStr := parts[1]
	monthsStr := ""
	if len(parts) > 2 {
		monthsStr = parts[2]
	}
	days, err := parseDays(strings.Split(daysStr, ","))
	if err != nil {
		return "", err
	}
	months := make(map[int]bool)
	if monthsStr != "" {
		months, err = parseMonths(strings.Split(monthsStr, ","))
		if err != nil {
			return "", err
		}
	}
	current := date

	for {
		current = current.AddDate(0, 0, 1)
		if len(months) > 0 {
			month := int(current.Month())
			if !months[month] {
				continue
			}
		}
		day := current.Day()
		lastDay := getLastDayOfMonth(current.Year(), int(current.Month()))
		if days[-1] && day == lastDay {
			if afterNow(current, now) {
				return current.Format(DateFormat), nil
			}
			continue
		}
		if days[-2] && day == lastDay-1 {
			if afterNow(current, now) {
				return current.Format(DateFormat), nil
			}
			continue
		}
		if days[day] {
			if afterNow(current, now) {
				return current.Format(DateFormat), nil
			}
		}
	}
}

// getLastDayOfMonth вычисление последнего дня месяца
func getLastDayOfMonth(year, month int) int {
	firstOfNextMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstOfNextMonth.AddDate(0, 0, -1)
	return lastDay.Day()
}

// parseMonths размечает месяцы используемые в правиле повторения
func parseMonths(monthsStr []string) (map[int]bool, error) {
	months := make(map[int]bool)
	for _, monthStr := range monthsStr {
		month, err := strconv.Atoi(monthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid month: %s", monthStr)
		}
		if month < 1 || month > 12 {
			return nil, fmt.Errorf("month must be between 1 and 12")
		}
		months[month] = true
	}
	return months, nil
}

// parseDays размечает дни используемые в месячном правиле повторения
func parseDays(daysStr []string) (map[int]bool, error) {
	days := make(map[int]bool)
	for _, dayStr := range daysStr {
		dayStr = strings.TrimSpace(dayStr)
		if dayStr == "-1" {
			days[-1] = true
			continue
		}
		if dayStr == "-2" {
			days[-2] = true
			continue
		}
		day, err := strconv.Atoi(dayStr)
		if err != nil {
			return nil, fmt.Errorf("invalid day: %s", dayStr)
		}
		if day < 1 || day > 31 {
			return nil, fmt.Errorf("day must be between 1 and 31")
		}
		days[day] = true
	}
	return days, nil
}

// handleDailyRule обработка интервального правила повторения (по дням)
func handleDailyRule(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", fmt.Errorf("missing interval for daily rule")
	}

	interval, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid interval %v", err)
	}

	if interval < 1 || interval > 400 {
		return "", fmt.Errorf("interval should be between 1 and 400")
	}

	result := date

	for {
		result = result.AddDate(0, 0, interval)
		if afterNow(result, now) {
			return result.Format(DateFormat), nil
		}
	}
}

// handleWeeklyRule обработка еженедельного правила повторения
func handleWeeklyRule(now time.Time, date time.Time, parts []string) (string, error) {
	if len(parts) < 2 {
		return "", fmt.Errorf("missing interval for weekly rule")
	}

	daysStr := strings.Split(parts[1], ",")
	weekDays := make(map[int]bool)
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(strings.TrimSpace(dayStr))
		if err != nil {
			return "", fmt.Errorf("invalid day: %s", dayStr)
		}
		if day < 1 || day > 7 {
			return "", fmt.Errorf("days must be in range from 1 to 7")
		}
		weekDays[day] = true
	}

	current := date

	for {
		current = current.AddDate(0, 0, 1)
		weekday := int(current.Weekday())

		if weekday == 0 {
			weekday = 7
		}

		if weekDays[weekday] && afterNow(current, now) {
			return current.Format(DateFormat), nil
		}
	}
}

// handleYearlyRule обработка ежегодного правила повторения
func handleYearlyRule(now time.Time, date time.Time) (string, error) {
	result := date

	for {
		result = result.AddDate(1, 0, 0)
		if afterNow(result, now) {
			return result.Format(DateFormat), nil
		}
	}
}

// afterNow проверяет что одна дата больше другой (игнорирует время)
func afterNow(date, now time.Time) bool {
	dateNorm := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	nowNorm := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	return dateNorm.After(nowNorm)
}

