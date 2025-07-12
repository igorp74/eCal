package main

import (
    "fmt"
    "regexp"
    "sort"
    "strings"
    "time"
)

// Helper struct to hold both foreground and background colors for display
type EventDisplayColors struct {
    FgColor string
    BgColor string
}

// pluralS returns "s" if count is not 1 or -1, otherwise returns an empty string.
func pluralS(count int) string {
    if count == 1 || count == -1 {
        return ""
    }
    return "s"
}

// ansiRegex is pre-compiled for efficiency to remove common ANSI escape codes.
// This simpler regex targets the most common SGR (Select Graphic Rendition) codes.
var ansiRegex = regexp.MustCompile("\x1b\\[[0-9;]*m")

// removeANSI removes ANSI escape codes from a string to calculate its visible length.
func removeANSI(s string) string {
    return ansiRegex.ReplaceAllString(s, "")
}

// GetMonthViewLines returns the lines for a single month's calendar view as a slice of strings.
func GetMonthViewLines(cfg Config, displayMonth time.Month, displayYear int, allEvents []Event) []string {
    var lines []string
    firstOfMonth := time.Date(displayYear, displayMonth, 1, 0, 0, 0, 0, cfg.TargetTime.Location()) // Use target time's location for consistency
    lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
    today := cfg.TargetTime // Use cfg.TargetTime for "today" consistency

    // Define the consistent visible width for the content area of a single month block.
    // Week Number Column (4 chars: " Wk " or " 22 ") + 7 Days (7 * 3 chars: " Mo ") = 4 + 21 = 25 visible characters.
    const monthBlockActualVisibleWidth = 25

    // Month/Year Header
    monthYearHeaderStr := fmt.Sprintf("%s %d", displayMonth.String(), displayYear)
    // Calculate padding needed to center the month/year string within monthBlockActualVisibleWidth
    paddingNeeded := monthBlockActualVisibleWidth - len(monthYearHeaderStr) // Calculate based on visible text length
    leftPadding := paddingNeeded / 2
    rightPadding := paddingNeeded - leftPadding
    // Construct the header line, ensuring bold style and padding are applied correctly
    centeredMonthYearHeader := fmt.Sprintf("%s%s%s%s%s", style_bold,
        strings.Repeat(" ", leftPadding),
        monthYearHeaderStr,
        strings.Repeat(" ", rightPadding),
        style_reset)
    lines = append(lines, centeredMonthYearHeader)

    var daysHeader []string
    if cfg.MondayFirst {
        daysHeader = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
    } else {
        daysHeader = []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
    }

    headerLine := ""
    // Week number column header (4 visible characters)
    if cfg.ShowWeekNum {
        headerLine += fmt.Sprintf("%s%3s%s ", fg_blue, "Wk", style_reset)
    } else {
        headerLine += strings.Repeat(" ", 4) // Always 4 spaces for week number column alignment
    }
    // Day headers (7 * 3 = 21 visible characters)
    for _, h := range daysHeader {
        headerLine += fmt.Sprintf("%-3s", h) // Each day header takes 3 visible spaces
    }
    // Pad the header line to ensure its visible length matches monthBlockActualVisibleWidth
    visibleHeaderLen := len(removeANSI(headerLine))
    headerLine += strings.Repeat(" ", monthBlockActualVisibleWidth-visibleHeaderLen)
    lines = append(lines, headerLine) // Removed TrimRight

    // Create a map to store unique event dates and their display colors for the current month.
    uniqueEventDatesForHighlight := make(map[time.Time]EventDisplayColors)
    for _, ev := range allEvents {
        // Only consider events within the current display month and year
        if ev.Date.Year() == displayYear && ev.Date.Month() == displayMonth {
            // Normalize event date to the same location as the calendar's current date
            dateOnly := time.Date(ev.Date.Year(), ev.Date.Month(), ev.Date.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())
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
        // Print week number column (4 visible characters)
        if cfg.ShowWeekNum {
            dayForWeekCalc := 1
            if weekRow == 0 {
                dayForWeekCalc = 1 // Use 1st of month for the first row (or a reasonable day if month starts mid-week)
            } else {
                // A day approximately in this row for ISO week calculation
                approxDayInRow := (weekRow * 7) + 1 - startDayOffset
                if approxDayInRow <= 0 {
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
                rowStr += fmt.Sprintf("%s%3d%s ", fg_blue, weekNo, style_reset) // Format to 3 chars, plus 1 space -> 4 visible chars
            } else {
                rowStr += strings.Repeat(" ", 4) // Padding for wk num column if empty
            }
        } else {
            rowStr += strings.Repeat(" ", 4) // Always 4 spaces for week number column alignment
        }

        hasDaysInRow := false
        for d := range 7 { // Iterate through 7 days of the week
            if weekRow == 0 && d < startDayOffset {
                rowStr += strings.Repeat(" ", 3) // Padding for days before the 1st of the month (3 visible chars)
            } else if currentDay > lastOfMonth.Day() {
                rowStr += strings.Repeat(" ", 3) // Padding for days after the last of the month (3 visible chars)
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

                eventDisplayColors, isEventDay := uniqueEventDatesForHighlight[currentDate]
                dayStr := fmt.Sprintf("%2d", dayToPrint) // Format to 2 characters, e.g., " 1", "10"

                // Apply colors/styles
                coloredDayStr := dayStr
                if isToday {
                    coloredDayStr = fmt.Sprintf("%s%s%s%s", fg_black, bg_yellow, dayStr, style_reset) // Highlight today's date
                } else if isEventDay && !isWeekend { // If it's an event day and not weekend AND we have colors
                    colorCodes := eventDisplayColors.FgColor + eventDisplayColors.BgColor
                    coloredDayStr = fmt.Sprintf("%s%s%s%s", colorCodes, style_bold, dayStr, style_reset)
                } else if isEventDay && isWeekend { // If it's an event day AND it is weekend
                    coloredDayStr = fmt.Sprintf("%s%s%s%s", style_bold, fg_red, dayStr, style_reset)
                } else if isWeekend {
                    coloredDayStr = fmt.Sprintf("%s%s%s", fg_red, dayStr, style_reset) // Just red for weekend
                }

                // Calculate visible length and pad explicitly to ensure each day block is 3 characters wide
                visibleLen := len(removeANSI(coloredDayStr))
                rowStr += coloredDayStr + strings.Repeat(" ", 3-visibleLen) // Each day block takes 3 visible spaces
                currentDay++
            }
        }

        // Pad the entire row to ensure its visible length matches monthBlockActualVisibleWidth
        visibleRowLen := len(removeANSI(rowStr))
        rowStr += strings.Repeat(" ", monthBlockActualVisibleWidth-visibleRowLen)
        lines = append(lines, rowStr) // Removed TrimRight

        // Break condition: if no days from the current month were printed in this row,
        // and we are past the first day of the month (currentDay > 1), then we are done.
        if !hasDaysInRow && currentDay > lastOfMonth.Day() {
            break
        }
        if currentDay > lastOfMonth.Day() && (startDayOffset+lastOfMonth.Day()) <= (weekRow+1)*7 {
            break // All days printed
        }
        if weekRow > 5 { // Safety break after 6 rows (max for a month)
            break
        }
    }
    return lines
}

// PrintCalendar renders multiple monthly calendars.
func PrintCalendar(cfg Config, startMonth time.Month, startYear int, allEvents []Event) {
    // Constants for layout
    const monthsPerRow = 3 // Number of months to display side-by-side
    // Use the same content width as GetMonthViewLines
    const monthBlockActualVisibleWidth = 25
    const interCalendarSpace = 4 // Spaces between each calendar block in a row

    var allMonthLines [][]string // Stores lines for each month: allMonthLines[monthIdx][lineIdx]

    // Generate lines for each individual month
    for i := 0; i < cfg.NumMonths; i++ {
        currentMonth := time.Month((int(startMonth)-1+i)%12 + 1)
        currentYear := startYear + ((int(startMonth)-1+i)/12)
        monthLines := GetMonthViewLines(cfg, currentMonth, currentYear, allEvents)
        allMonthLines = append(allMonthLines, monthLines)
    }

    // Determine max height among all months for consistent alignment
    maxHeight := 0
    for _, lines := range allMonthLines {
        if len(lines) > maxHeight {
            maxHeight = len(lines)
        }
    }

    // Ensure all month views have the same height by padding with empty strings
    for i := range allMonthLines {
        for len(allMonthLines[i]) < maxHeight {
            // Pad with spaces matching the monthBlockActualVisibleWidth
            allMonthLines[i] = append(allMonthLines[i], strings.Repeat(" ", monthBlockActualVisibleWidth))
        }
    }

    // fmt.Println() // Add an empty line before print calendar(s)
    // Print months in rows of `monthsPerRow`
    for i := 0; i < cfg.NumMonths; i += monthsPerRow {
        for lineIdx := 0; lineIdx < maxHeight; lineIdx++ {
            rowOutput := ""
            for j := range monthsPerRow {
                monthIdx := i + j
                if monthIdx < cfg.NumMonths {
                    line := allMonthLines[monthIdx][lineIdx]
                    rowOutput += line
                    // Add inter-calendar spacing, but not after the last month in the row
                    if j < monthsPerRow-1 {
                        rowOutput += strings.Repeat(" ", interCalendarSpace)
                    }
                } else {
                    // If fewer than monthsPerRow months in the last row, fill with spaces
                    rowOutput += strings.Repeat(" ", monthBlockActualVisibleWidth)
                    if j < monthsPerRow-1 {
                        rowOutput += strings.Repeat(" ", interCalendarSpace)
                    }
                }
            }
            fmt.Println(rowOutput)
        }
        fmt.Println() // Add an empty line between rows of months
    }
}

// PrintEventList renders a combined event list for the displayed period.
func PrintEventList(cfg Config, startMonth time.Month, startYear int, allEvents []Event) {
    fmt.Printf("%sEvents for displayed period:%s\n", style_bold, style_reset)
    foundEvents := false
    uniqueEventsForList := make(map[string]Event)

    // Determine the end month and year of the display range
    endMonth := time.Month((int(startMonth)-1+cfg.NumMonths-1)%12 + 1)
    endYear := startYear + ((int(startMonth)-1+cfg.NumMonths-1)/12)

    startDate := time.Date(startYear, startMonth, 1, 0, 0, 0, 0, cfg.TargetTime.Location())
    endDate := time.Date(endYear, endMonth, 1, 0, 0, 0, 0, cfg.TargetTime.Location()).AddDate(0, 1, -1) // Last day of the end month

    for _, e := range allEvents {
        eventDate := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())

        if (eventDate.Equal(startDate) || eventDate.After(startDate)) && (eventDate.Equal(endDate) || eventDate.Before(endDate)) {
            dateOnlyStr := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, e.Date.Location()).Format("2006-01-02")
            compositeKey := dateOnlyStr + "::" + e.Description
            if _, exists := uniqueEventsForList[compositeKey]; !exists {
                uniqueEventsForList[compositeKey] = e
            }
        }
    }

    var sortedUniqueEvents []Event
    for _, ev := range uniqueEventsForList {
        sortedUniqueEvents = append(sortedUniqueEvents, ev)
    }

    // Sort the events by date
    sort.Slice(sortedUniqueEvents, func(i, j int) bool {
        return sortedUniqueEvents[i].Date.Before(sortedUniqueEvents[j].Date)
    })

    if len(sortedUniqueEvents) > 0 {
        foundEvents = true
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

            // Use explicit emoji if provided, otherwise fall back to default based on type
            displayEmoji := e.Emoji
            if displayEmoji == "" {
                displayEmoji = getDefaultEmoji(e.Type)
            }

            // Use e.DisplayColor and e.DisplayBgColor for the event list output as well
            fmt.Printf("%s%s%2d%s %s, %s %s: %s %s", e.DisplayColor, e.DisplayBgColor, e.Date.Day(), daySuffix, e.Date.Month().String()[:3], e.Date.Weekday().String()[:3], style_reset, displayEmoji, e.Description)

            if e.IsBirthday && !e.BirthDate.IsZero() {
                age := cfg.TargetTime.Year() - e.BirthDate.Year()
                if cfg.TargetTime.Month() < e.BirthDate.Month() ||
                    (cfg.TargetTime.Month() == e.BirthDate.Month() && cfg.TargetTime.Day() < e.BirthDate.Day()) {
                    age--
                }
                if age >= 0 {
                    fmt.Printf(" (Age: %d)", age)
                }
            }

            eventDayStart := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())
            todayStart := time.Date(cfg.TargetTime.Year(), cfg.TargetTime.Month(), cfg.TargetTime.Day(), 0, 0, 0, 0, cfg.TargetTime.Location())
            daysDiff := int(eventDayStart.Sub(todayStart).Hours() / 24)

            if daysDiff == 0 {
                fmt.Printf(" %s(Today)%s", fg_blue, style_reset)
            } else if daysDiff > 0 {
                fmt.Printf(" %s(In %s%d%s%s day%s)%s", fg_green, style_bold, daysDiff, style_reset, fg_green, pluralS(daysDiff), style_reset)
            } else {
                fmt.Printf(" %s(%d day%s ago)%s", fg_blue, -daysDiff, pluralS(-daysDiff), style_reset)
            }
            fmt.Println()
        }
    }

    if !foundEvents {
        fmt.Println("No events in the displayed period.")
    }
}

