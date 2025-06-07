package main

import (
    "fmt"
    "strings"
    "time"
)

// Helper struct to hold both foreground and background colors for display
type EventDisplayColors struct {
    FgColor string
    BgColor string
}

func pluralS(count int) string {
    if count == 1 || count == -1 {
        return ""
    }
    return "s"
}

// PrintCalendar renders the monthly calendar to the console.
func PrintCalendar(cfg Config, displayMonth time.Month, displayYear int, allEvents []Event) {
    firstOfMonth := time.Date(displayYear, displayMonth, 1, 0, 0, 0, 0, cfg.TargetTime.Location()) // Use target time's location for consistency
    lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
    today := time.Now()

    fmt.Printf("\n%s%s%s %d%s\n", style_bold, strings.Repeat(" ", (28-len(displayMonth.String())-len(fmt.Sprintf("%d", displayYear)))/2), displayMonth.String(), displayYear, style_reset) // Centered month/year

    var daysHeader []string
    if cfg.MondayFirst {
        daysHeader = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
    } else {
        daysHeader = []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
    }

    headerLine := ""
    if cfg.ShowWeekNum {
        headerLine += fmt.Sprintf("%s%3s%s ", fg_blue, "Wk", style_reset)
    }
    for _, h := range daysHeader {
        headerLine += fmt.Sprintf("%-3s", h) // Each day header takes 3 spaces (e.g. "Mo ")
    }
    fmt.Println(strings.TrimRight(headerLine, " "))

    // Create a map to store unique event dates and their display colors for the current month.
    // This ensures a date is highlighted only once and with its specific color if defined.
    uniqueEventDatesForHighlight := make(map[time.Time]EventDisplayColors) // Changed value type to EventDisplayColors struct
    for _, ev := range allEvents {
        // Only consider events within the current display month and year
        if ev.Date.Year() == displayYear && ev.Date.Month() == displayMonth {
            // Normalize event date to the same location as the calendar's current date
            // This is CRUCIAL for map key consistency.
            dateOnly := time.Date(ev.Date.Year(), ev.Date.Month(), ev.Date.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())
            // Store the event's display colors. If multiple events on same day, first one wins.
            // Or you could implement a priority system if needed.
            if _, exists := uniqueEventDatesForHighlight[dateOnly]; !exists {
                uniqueEventDatesForHighlight[dateOnly] = EventDisplayColors{
                    FgColor: ev.DisplayColor,
                    BgColor: ev.DisplayBgColor,
                }
            }
        }
    }

    // Calculate starting weekday offset
    startDayOffset := int(firstOfMonth.Weekday()) // Sunday = 0, ..., Saturday = 6
    if cfg.MondayFirst {
        startDayOffset = (startDayOffset - 1 + 7) % 7 // Monday = 0, ..., Sunday = 6
    }

    currentDay := 1
    for weekRow := 0; ; weekRow++ {
        rowStr := ""
        // Print week number
        if cfg.ShowWeekNum {
            // Calculate week number based on a day in the current row
            dayForWeekCalc := 1
            if weekRow == 0 {
                dayForWeekCalc = 1 // Use 1st of month for the first row
            } else {
                // A day approximately in this row
                approxDayInRow := (weekRow * 7) + 1 - startDayOffset
                if approxDayInRow <= 0 { // Should not happen if logic is correct
                    dayForWeekCalc = 1
                } else if approxDayInRow > lastOfMonth.Day() {
                    dayForWeekCalc = lastOfMonth.Day()
                } else {
                    dayForWeekCalc = approxDayInRow
                }
            }
            // Only print week number if there are actual days from the month in this row or previous rows had days
            if currentDay <= lastOfMonth.Day() || (weekRow > 0 && (startDayOffset+lastOfMonth.Day()) > (weekRow*7)) {
                _, weekNo := time.Date(displayYear, displayMonth, dayForWeekCalc, 0, 0, 0, 0, cfg.TargetTime.Location()).ISOWeek()
                rowStr += fmt.Sprintf("%s%3d%s ", fg_blue, weekNo, style_reset)
            } else {
                rowStr += fmt.Sprintf("%4s", "") // Padding for wk num column
            }
        }

        hasDaysInRow := false
        for d := range 7 {
            if weekRow == 0 && d < startDayOffset {
                rowStr += fmt.Sprintf("%-3s", "") // Padding for days before the 1st
            } else if currentDay > lastOfMonth.Day() {
                rowStr += fmt.Sprintf("%-3s", "") // Padding for days after the last
            } else {
                hasDaysInRow = true
                dayToPrint := currentDay

                isWeekend := false
                currentDate := time.Date(displayYear, displayMonth, dayToPrint, 0, 0, 0, 0, cfg.TargetTime.Location())
                isToday := currentDate.Year() == today.Year() && currentDate.Month() == today.Month() && currentDate.Day() == today.Day()
                weekday := currentDate.Weekday()

                if cfg.MondayFirst {
                    if weekday == time.Saturday || weekday == time.Sunday {
                        isWeekend = true
                    }
                } else { // Sunday first
                    if weekday == time.Sunday || weekday == time.Saturday {
                        isWeekend = true
                    }
                }

                // Get the event colors from the map.
                eventDisplayColors, isEventDay := uniqueEventDatesForHighlight[currentDate]

                dayStr := fmt.Sprintf("%2d", dayToPrint)
                dayFormatted := ""
                if isToday {
                    dayFormatted = fmt.Sprintf("%s%s%s%s ", fg_black, bg_yellow, dayStr, style_reset) // Highlight today's date
                } else if isEventDay && !isWeekend{ // If it's an event day and not weekend AND we have colors
                    // Apply both foreground and background colors if present
                    colorCodes := eventDisplayColors.FgColor + eventDisplayColors.BgColor
                    dayFormatted = fmt.Sprintf("%s%s%-3s%s", colorCodes, style_bold, dayStr, style_reset)
                } else if isEventDay && isWeekend { // If it's an event day AND it is weekend
                    // Apply both foreground and background colors if present
                    dayFormatted = fmt.Sprintf("%s%s%-3s%s", style_bold, fg_red, dayStr, style_reset)
                } else if isWeekend {
                    dayFormatted = fmt.Sprintf("%s%-3s%s", fg_red, dayStr, style_reset)
                } else {
                    dayFormatted = dayStr
                }
                rowStr += fmt.Sprintf("%-3s", dayFormatted) // Each day takes 3 spaces
                currentDay++
            }
        }
        fmt.Println(strings.TrimRight(rowStr, " "))

        // Break condition: if no days from the current month were printed in this row,
        // and we are past the first day of the month (currentDay > 1), then we are done.
        if !hasDaysInRow && currentDay > lastOfMonth.Day() {
            break
        }
        if currentDay > lastOfMonth.Day() && (startDayOffset+lastOfMonth.Day()) <= (weekRow+1)*7 {
            break // All days printed
        }
        if weekRow > 5 {
            break
        } // Safety break after 6 rows (max for a month)
    }

    // --- Event List Uniqueness Enhancement ---
    fmt.Printf("\n%sEvents for %s %d:%s\n", style_bold, displayMonth, displayYear, style_reset)
    foundEventsThisMonth := false

    // Use a map to store unique events by date AND description for the list below the calendar
    // The key will be a string combining date and description to ensure uniqueness.
    uniqueEventsForList := make(map[string]Event)
    for _, e := range allEvents {
        if e.Date.Year() == displayYear && e.Date.Month() == displayMonth {
            // Normalize event date to remove time components for comparison, and use its string representation.
            dateOnlyStr := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, e.Date.Location()).Format("2006-01-02")
            // Create a composite key with date and description
            compositeKey := dateOnlyStr + "::" + e.Description // Using "::" as a separator to minimize collision risk

            if _, exists := uniqueEventsForList[compositeKey]; !exists {
                uniqueEventsForList[compositeKey] = e
            }
        }
    }

    // To print in chronological order, convert the map to a slice and sort it
    var sortedUniqueEvents []Event
    for _, ev := range uniqueEventsForList {
        sortedUniqueEvents = append(sortedUniqueEvents, ev)
    }

    // Sort the events by date
    // This is a simple bubble sort for demonstration; for large lists, use sort.Slice
    for i := range len(sortedUniqueEvents)-1 {
        for j := range len(sortedUniqueEvents)-i-1 {
            if sortedUniqueEvents[j].Date.After(sortedUniqueEvents[j+1].Date) {
                sortedUniqueEvents[j], sortedUniqueEvents[j+1] = sortedUniqueEvents[j+1], sortedUniqueEvents[j]
            }
        }
    }

    if len(sortedUniqueEvents) > 0 {
        foundEventsThisMonth = true
        for _, e := range sortedUniqueEvents {
            daySuffix := "th"
            switch e.Date.Day() {
            case 1, 21, 31:
                daySuffix = "st"
            case 2, 22:
                daySuffix = "nd"
            case 3, 23:
                daySuffix = "rd"
            }
            emoji := getEmoji(e.Type)

            // Use e.DisplayColor and e.DisplayBgColor for the event list output as well
            fmt.Printf("%s%s%s%2d%s %s%s: %s %s", e.DisplayColor, e.DisplayBgColor, style_bold, e.Date.Day(), daySuffix, e.Date.Weekday().String()[:3], style_reset, emoji, e.Description)

            if e.IsBirthday && !e.BirthDate.IsZero() {
                // Calculate age based on cfg.TargetTime (current time when app runs)
                age := cfg.TargetTime.Year() - e.BirthDate.Year()
                // Adjust if birthday hasn't occurred yet this year relative to TargetTime
                if cfg.TargetTime.Month() < e.BirthDate.Month() ||
                    (cfg.TargetTime.Month() == e.BirthDate.Month() && cfg.TargetTime.Day() < e.BirthDate.Day()) {
                    age--
                }
                if age >= 0 { // Only show age if it's sensible
                    fmt.Printf(" (Age: %d)", age)
                }
            }

            // Calculate days before/after relative to cfg.TargetTime
            // For comparison, ensure dates are at midnight in the same location
            eventDayStart := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())
            todayStart := time.Date(cfg.TargetTime.Year(), cfg.TargetTime.Month(), cfg.TargetTime.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())

            daysDiff := int(eventDayStart.Sub(todayStart).Hours() / 24)

            if daysDiff == 0 {
                fmt.Printf(" %s(Today)%s", fg_blue, style_reset)
            } else if daysDiff > 0 {
                fmt.Printf(" %s(In %d day%s)%s", fg_blue, daysDiff, pluralS(daysDiff), style_reset)
            } else { // daysDiff < 0
                fmt.Printf(" %s(%d day%s ago)%s", fg_blue, -daysDiff, pluralS(-daysDiff), style_reset)
            }
            fmt.Println()
        }
    }

    if !foundEventsThisMonth {
        fmt.Println("No events this month.")
    }
    fmt.Println()
}
