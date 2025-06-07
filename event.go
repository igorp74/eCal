package main

import (
    "bufio"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

var (
    // MM/DOW#N (e.g., 5/1#1 for 1st Monday of May; DOW: 1=Mon, ..., 7=Sun)
    reNthWeekday = regexp.MustCompile(`^(\d{1,2})/([1-7])#([1-5])$`)
    // E or E+N or E-N
    reEaster = regexp.MustCompile(`^E([+-]?)(\d*)$`)
    // MM/DD or MM/DD? or MM/DD?YYYY or MM/DD?D[+-]N (D is 0-6 for Sun-Sat)
    reMonthDay = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})(\?(?:(\d{4})|([0-6][+-]\d+)|))?$`)
    // MM/DD/YYYY
    reUsDate = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)
    // DD-MM-YYYY
    reIsoDate = regexp.MustCompile(`^(\d{1,2})-(\d{1,2})-(\d{4})$`)

    // Modified regex to capture:
    // Group 1: event type (e.g., "ie", "birthday")
    // Group 2: optional foreground color name (e.g., "red", "magenta") if present after comma
    // Group 3: optional background color name (e.g., "white", "blue") if present after second comma
    // Group 4: the rest of the description
    // It now handles: [type], [type,fg_color], or [type,fg_color,bg_color]
    reEventTypeAndColor = regexp.MustCompile(`^\s*\[\s*([^,\]]+?)(?:,\s*([^,\]]+?))?(?:,\s*([^\]]+?))?\s*\]\s*(.*)$`)
)

// parseEventDate attempts to parse a date string from an event file.
// yearContext is the year for which annual events should be resolved.
// Returns: parsedDate (for yearContext), isAnnual, isBirthdayCandidate, birthDate (if present), recurrenceRule, specificYearInRule, error
func parseEventDate(dateStr string, yearContext int) (time.Time, bool, bool, time.Time, string, bool, error) {
    var parsedDate, birthDateVal time.Time
    var isAnnual, isBirthdayCandidate, specificYearInRule bool
    var recurrenceRule string

    // 1. Easter relative: E, E+N, E-N
    if matches := reEaster.FindStringSubmatch(dateStr); len(matches) > 0 {
        easterD := CalculateEaster(yearContext)
        offset := 0
        if matches[2] != "" {
            offset, _ = strconv.Atoi(matches[2])
        }
        if matches[1] == "-" {
            offset = -offset
        }
        parsedDate = easterD.AddDate(0, 0, offset)
        isAnnual = true // Easter events are annual relative to the given year's Easter
        recurrenceRule = dateStr
        return parsedDate, isAnnual, false, time.Time{}, recurrenceRule, false, nil
    }

    // 2. Nth DOW of Month: MM/DOW#N (DOW: 1=Mon .. 7=Sun, Nth: 1-5)
    if matches := reNthWeekday.FindStringSubmatch(dateStr); len(matches) > 0 {
        month, _ := strconv.Atoi(matches[1])
        dowUser, _ := strconv.Atoi(matches[2]) // 1 (Mon) to 7 (Sun)
        nth, _ := strconv.Atoi(matches[3])     // 1 to 5

        if month < 1 || month > 12 {
            return time.Time{}, false, false, time.Time{}, "", false, fmt.Errorf("invalid month in MM/DOW#N: %s", dateStr)
        }
        // User DOW (1=Mon..7=Sun) to time.Weekday (Sunday=0..Saturday=6)
        mapUserDowToStd := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}
        targetWeekday := mapUserDowToStd[dowUser-1]

        pDate, err := NthWeekdayOfMonth(yearContext, time.Month(month), nth, targetWeekday)
        if err != nil {
            return time.Time{}, false, false, time.Time{}, "", false, fmt.Errorf("calculating Nth weekday for %s: %w", dateStr, err)
        }
        parsedDate = pDate
        isAnnual = true
        recurrenceRule = dateStr
        return parsedDate, isAnnual, false, time.Time{}, recurrenceRule, false, nil
    }

    // 3. MM/DD based: MM/DD, MM/DD?, MM/DD?YYYY, MM/DD?D[+-]N
    if matches := reMonthDay.FindStringSubmatch(dateStr); len(matches) > 0 {
        month, _ := strconv.Atoi(matches[1])
        day, _ := strconv.Atoi(matches[2])
        yearToUse := yearContext
        isAnnual = true // Initially assume annual unless a specific year is found

        if month < 1 || month > 12 || day < 1 || day > 31 { // Basic validation
            return time.Time{}, false, false, time.Time{}, "", false, fmt.Errorf("invalid month/day in MM/DD rule: %s", dateStr)
        }

        optYearStr := matches[4]     // This part if present
        optCondRuleStr := matches[5] // D[+-]N part if present

        if optYearStr != "" { // ?YYYY
            if yr, errYr := strconv.Atoi(optYearStr); errYr == nil {
                yearToUse = yr
                isAnnual = false // Year is specified
                specificYearInRule = true
            }
        }

        baseDate := time.Date(yearToUse, time.Month(month), day, 0, 0, 0, 0, time.UTC)

        if optCondRuleStr != "" { // ?D[+-]N, e.g. 6+2 for "if on Sat, add 2 days"
            // D is 0-6 (Sun-Sat)
            condRuleRegex := regexp.MustCompile(`^([0-6])([+-])(\d+)$`)
            condMatches := condRuleRegex.FindStringSubmatch(optCondRuleStr)
            if len(condMatches) == 4 {
                dwVal, _ := strconv.Atoi(condMatches[1])
                op := condMatches[2]
                offsetVal, _ := strconv.Atoi(condMatches[3])

                conditionalWeekday := time.Weekday(dwVal) // 0=Sun, ..., 6=Sat
                if baseDate.Weekday() == conditionalWeekday {
                    if op == "-" {
                        offsetVal = -offsetVal
                    }
                    parsedDate = baseDate.AddDate(0, 0, offsetVal)
                } else {
                    parsedDate = baseDate
                }
                recurrenceRule = dateStr // The whole MM/DD?D[+-]N
                // isAnnual remains true as this rule is checked annually against yearContext's baseDate
                return parsedDate, isAnnual, false, time.Time{}, recurrenceRule, specificYearInRule, nil
            }
        }

        // If no conditional rule, or year not specified by ?YYYY
        parsedDate = baseDate
        // isAnnual is true if specificYearInRule is false. If specificYearInRule is true, isAnnual is false.
        isAnnual = !specificYearInRule
        return parsedDate, isAnnual, false, time.Time{}, dateStr, specificYearInRule, nil
    }

    // 4. US Date: MM/DD/YYYY
    if matches := reUsDate.FindStringSubmatch(dateStr); len(matches) > 0 {
        month, _ := strconv.Atoi(matches[1])
        day, _ := strconv.Atoi(matches[2])
        year, _ := strconv.Atoi(matches[3])
        if month < 1 || month > 12 || day < 1 || day > 31 {
            return time.Time{}, false, false, time.Time{}, "", false, fmt.Errorf("invalid month/day in MM/DD/YYYY: %s", dateStr)
        }
        parsedDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
        isAnnual = false
        isBirthdayCandidate = true
        birthDateVal = parsedDate
        specificYearInRule = true
        return parsedDate, isAnnual, isBirthdayCandidate, birthDateVal, "", specificYearInRule, nil
    }

    // 5. ISO-like Date: DD-MM-YYYY
    if matches := reIsoDate.FindStringSubmatch(dateStr); len(matches) > 0 {
        day, _ := strconv.Atoi(matches[1])
        month, _ := strconv.Atoi(matches[2])
        year, _ := strconv.Atoi(matches[3])
        if month < 1 || month > 12 || day < 1 || day > 31 {
            return time.Time{}, false, false, time.Time{}, "", false, fmt.Errorf("invalid month/day in DD-MM-YYYY: %s", dateStr)
        }
        parsedDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
        isAnnual = false
        isBirthdayCandidate = true
        birthDateVal = parsedDate
        specificYearInRule = true
        return parsedDate, isAnnual, isBirthdayCandidate, birthDateVal, "", specificYearInRule, nil
    }

    return time.Time{}, false, false, time.Time{}, "", false, fmt.Errorf("unknown date format: '%s'", dateStr)
}

// LoadEvents reads events from the specified file for a given year context.
func LoadEvents(filePath string, yearContext int) ([]Event, error) {
    file, err := os.Open(filePath)
    if err != nil {
        if os.IsNotExist(err) {
            fmt.Fprintf(os.Stderr, "Info: Events file '%s' not found. No events will be loaded.\n", filePath)
            return []Event{}, nil // No events file is not a critical error
        }
        return nil, fmt.Errorf("opening events file '%s': %w", filePath, err)
    }
    defer file.Close()

    var events []Event
    scanner := bufio.NewScanner(file)
    lineNumber := 0
    for scanner.Scan() {
        lineNumber++
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") { // Skip empty lines and comments
            continue
        }

        parts := strings.SplitN(line, ";", 2)
        if len(parts) != 2 {
            fmt.Fprintf(os.Stderr, "Warning (line %d): Malformed event (missing ';'): %s\n", lineNumber, line)
            continue
        }

        dateStr := strings.TrimSpace(parts[0])
        descPart := strings.TrimSpace(parts[1])

        var eventType, eventDesc, fgColor, bgColor string
        typeMatch := reEventTypeAndColor.FindStringSubmatch(descPart)
        if len(typeMatch) >= 5 { // Expect at least 5 groups: full match, type, fg_color, bg_color, description
            eventType = strings.TrimSpace(typeMatch[1])    // Group 1: event type
            fgColorName := strings.TrimSpace(typeMatch[2]) // Group 2: optional fg color name
            bgColorName := strings.TrimSpace(typeMatch[3]) // Group 3: optional bg color name
            eventDesc = strings.TrimSpace(typeMatch[4])    // Group 4: the rest of the description

            fgColor = GetFgColorCode(fgColorName)
            bgColor = GetBgColorCode(bgColorName)

        } else if len(typeMatch) >= 4 { // Handles [type,fg_color]
            eventType = strings.TrimSpace(typeMatch[1])
            fgColorName := strings.TrimSpace(typeMatch[2])
            eventDesc = strings.TrimSpace(typeMatch[3]) // Note: Group 3 is now the description here

            fgColor = GetFgColorCode(fgColorName)
            bgColor = "" // No background color specified
            
        } else if len(typeMatch) >= 2 { // Handles [type]
            eventType = strings.TrimSpace(typeMatch[1])
            eventDesc = strings.TrimSpace(typeMatch[2]) // Note: Group 2 is now the description here
            
            fgColor = fg_green // Default foreground
            bgColor = "" // No background color
        } else {
            // Fallback for cases where [type] or [type,color] is not matched at all
            eventDesc = descPart // No type or color found, use whole string as description
            eventType = "default"
            fgColor = fg_green // Default highlight color
            bgColor = ""     // Default to no background color
            fmt.Fprintf(os.Stderr, "Warning (line %d): Event description format unexpected, treating as plain description: %s\n", lineNumber, descPart)
        }

        parsedDate, isAnnual, isBdayCandidate, bDateVal, recRule, specYearRule, err := parseEventDate(dateStr, yearContext)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Warning (line %d): Skipping event due to date parse error ('%s'): %v\n", lineNumber, dateStr, err)
            continue
        }

        // Finalize event date and birthday status
        actualEventDate := parsedDate
        isActualBirthday := strings.ToLower(eventType) == "birthday" && isBdayCandidate && !bDateVal.IsZero()

        // If it's an annual event (not fixed to a specific year by its rule),
        // its date should be in the yearContext.
        // If it has a specific year in its rule (specYearRule=true), its date is fixed.
        // For birthdays, bDateVal holds the birth year. The event occurs annually.
        if isActualBirthday {
            // Birthday occurs on bDateVal.Month and bDateVal.Day in yearContext
            actualEventDate = time.Date(yearContext, bDateVal.Month(), bDateVal.Day(), 0, 0, 0, 0, time.UTC)
            isAnnual = true // Birthdays are effectively annual occurrences
        } else if isAnnual && actualEventDate.Year() != yearContext {
            // Ensure annual non-birthday events are set for the correct yearContext
            actualEventDate = time.Date(yearContext, actualEventDate.Month(), actualEventDate.Day(), 0, 0, 0, 0, time.UTC)
        } else if specYearRule && actualEventDate.Year() != yearContext {
            // If rule specified a year and it's not yearContext, and it's NOT a birthday, then skip.
            continue
        }

        event := Event{
            Date:             actualEventDate,
            OriginalDateStr:  dateStr,
            Description:      eventDesc,
            Type:             eventType,
            IsAnnual:         isAnnual,
            IsBirthday:       isActualBirthday,
            BirthDate:        bDateVal, // Store the original birth date
            RecurrenceRule:   recRule,
            SpecificYearRule: specYearRule,
            DisplayColor:     fgColor, // Store the determined foreground color
            DisplayBgColor:   bgColor, // Store the determined background color
        }
        events = append(events, event)
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("reading events file: %w", err)
    }
    return events, nil
}
