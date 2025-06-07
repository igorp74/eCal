package main

import (
    "fmt"
    "time"
)

// CalculateEaster calculates the date of Easter Sunday for a given year (Gregorian algorithm).
// Returns the date in UTC.
func CalculateEaster(year int) time.Time {
    // Anonymous Gregorian algorithm
    a := year % 19
    b := year / 100
    c := year % 100
    d := b / 4
    e := b % 4
    f := (b + 8) / 25
    g := (b - f + 1) / 3
    h := (19*a + b - d - g + 15) % 30
    i := c / 4
    k := c % 4
    l := (32 + 2*e + 2*i - h - k) % 7
    m := (a + 11*h + 22*l) / 451
    month := (h + l - 7*m + 114) / 31
    day := ((h + l - 7*m + 114) % 31) + 1
    return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// NthWeekdayOfMonth calculates the date of the Nth specific weekday in a given month and year.
// nth: 1 for 1st, 2 for 2nd, etc. (1-5)
// targetWeekday: time.Weekday (Sunday=0, ..., Saturday=6)
// Returns the date in UTC.
func NthWeekdayOfMonth(year int, month time.Month, nth int, targetWeekday time.Weekday) (time.Time, error) {
    if nth <= 0 || nth > 5 { // A month can have at most 5 occurrences of a specific weekday
        return time.Time{}, fmt.Errorf("invalid 'nth' value: %d, must be between 1 and 5", nth)
    }

    firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
    weekdayOfFirst := firstOfMonth.Weekday()

    // Days to add to the 1st of the month to get to the first targetWeekday
    daysToAdd := (int(targetWeekday) - int(weekdayOfFirst) + 7) % 7

    // Date of the first occurrence of the targetWeekday
    firstOccurrenceDate := firstOfMonth.AddDate(0, 0, daysToAdd)

    // Date of the Nth occurrence
    nthOccurrenceDate := firstOccurrenceDate.AddDate(0, 0, (nth-1)*7)

    // Check if the Nth occurrence is still within the same month
    if nthOccurrenceDate.Month() != month {
        // Correction for 5th occurance to be the last DOW in the month
        return firstOccurrenceDate.AddDate(0, 0, (nth-2)*7), nil
        // More strict approach - if there is no 5th occurence, it will throw the error
        // return time.Time{}, fmt.Errorf("%d(th/st/nd/rd) %s not found in %s %d", nth, targetWeekday, month, year)
    }

    return nthOccurrenceDate, nil
}

// GetFirstDayOfISOWeek returns the date of Monday of the given ISO week and year.
// Returns the date in UTC.
func GetFirstDayOfISOWeek(year, week int) (time.Time, error) {
    if week < 1 || week > 53 {
        return time.Time{}, fmt.Errorf("invalid week number: %d", week)
    }

    // A more robust way:
    // Start from Jan 1st of the target year.
    // Iterate forward until we find a day that belongs to the target ISO week.
    // Then, go back to the Monday of that week.
    dateInTargetWeek := time.Date(year, time.January, 1, 0,0,0,0, time.UTC)
    for i:= range 370 { // Iterate through days (max a bit over a year)
    // for i:=0; i<370; i++ { // Iterate through days (max a bit over a year)
        y, w := dateInTargetWeek.ISOWeek()
        if y == year && w == week {
            break // Found a day in the target week
        }
        if y > year || (y == year && w > week && week !=1) { // Overshot, or wrapped around for week 1 next year
             if week == 1 && y == year+1 && w == 1 { break } // week 1 of next year is fine
             return time.Time{}, fmt.Errorf("could not accurately find a day in week %d of year %d (overshot)", week, year)
        }
        dateInTargetWeek = dateInTargetWeek.AddDate(0,0,1)
        if i == 369 {
            return time.Time{}, fmt.Errorf("could not find a day in week %d of year %d after 370 attempts", week, year)
        }
    }

    // Now dateInTargetWeek is a date within the desired ISO week.
    // Find the Monday of this week.
    wd := dateInTargetWeek.Weekday()
    offsetToMonday := 0
    if wd == time.Sunday { // Sunday is 0
        offsetToMonday = -6 // Go back 6 days
    } else {
        offsetToMonday = int(time.Monday - wd) // e.g. if wd is Wednesday (3), Monday-Wednesday = 1-3 = -2. Add -2 days.
    }

    firstDay := dateInTargetWeek.AddDate(0, 0, offsetToMonday)
    // Final check
    fy, fw := firstDay.ISOWeek()
    if fy == year && fw == week {
        return firstDay, nil
    }
    // If firstDay falls into the previous year (e.g. week 1 of 2024 might start in Dec 2023)
    if fw == week && (fy == year || (fy == year -1 && week >= 52)) {
        return firstDay, nil
    }


    return time.Time{}, fmt.Errorf("could not determine first day of ISO week %d for year %d (final check failed: %s, %d, %d)", week, year, firstDay.Format("2006-01-02"), fy, fw)
}
