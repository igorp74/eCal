package main

import (
    "flag"
    "fmt"
    "os"
    "time"
)

func main() {
    currentTime := time.Now() // Use system's current time

    // Default configuration
    cfg := Config{
        Year:        currentTime.Year(),
        Month:       currentTime.Month(),
        Week:        0, // Will be checked if > 0
        MondayFirst: true,
        EventsFile:  "events.txt", // Default events file name
        ShowWeekNum: true,
        TargetTime:  currentTime, // Reference time for age/countdown
        NumMonths:   1,           // Default to showing 1 month
        NumColumns:  3,
        DisplayMode: DisplayBoth, // Default to showing both calendar and events
    }

    // Command-line flags
    yearFlag    := flag.Int("y",  0, "Year for the calendar (default: current year). Also used with -week.")
    monthFlag   := flag.Int("m",  0, "Month for the calendar (1-12) (default: current month).")
    weekFlag    := flag.Int("w",  0, "Week number for the calendar (1-53). If used with -year, overrides -month.")
    monthsFlag  := flag.Int("mn", 1, "Number of months to display (1, 3, 6, or 12).")
    columnsFlag := flag.Int("c",  3, "Number of columns to display (1, 3, 4, 6, or 12).")
    displayFlag := flag.String("d", DisplayBoth, "What to display: 'calendar', 'events', or 'both' (default).") // New display flag

    flag.BoolVar(&cfg.MondayFirst,  "monday", cfg.MondayFirst, "Set Monday as the first day of the week.")
    flag.StringVar(&cfg.EventsFile, "f",      cfg.EventsFile,  "Path to the events file.")
    flag.BoolVar(&cfg.ShowWeekNum,  "wk",     cfg.ShowWeekNum, "Show week numbers.")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        fmt.Fprintf(os.Stderr, " %s [options]\n\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
        fmt.Fprintf(os.Stderr, "\n\033[1mExamples:\033[0m\n")
        fmt.Fprintf(os.Stderr, "  %s -y 2024 -m 12\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s -y 2024 -w 50\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s -f my_holidays.txt -monday\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s -m 7 -y 2025 -mn 3\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s -d calendar\n", os.Args[0]) // New example
        fmt.Fprintf(os.Stderr, "  %s -d events -f my_events.txt\n", os.Args[0]) // New example
    }
    flag.Parse()

    // Process flags
    if *yearFlag != 0 {
        cfg.Year = *yearFlag
    }
    if *monthFlag != 0 {
        if *monthFlag >= 1 && *monthFlag <= 12 {
            cfg.Month = time.Month(*monthFlag)
        } else {
            fmt.Fprintf(os.Stderr, "Error: Invalid month value %d. Must be between 1 and 12.\n", *monthFlag)
            flag.Usage()
            os.Exit(1)
        }
    }
    if *weekFlag != 0 {
        if *weekFlag >= 1 && *weekFlag <= 53 {
            cfg.Week = *weekFlag
            if *yearFlag == 0 { // -week requires -year
                fmt.Fprintf(os.Stderr, "Error: -week flag requires -year to be specified.\n")
                flag.Usage()
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "Error: Invalid week value %d. Must be between 1 and 53.\n", *weekFlag)
            flag.Usage()
            os.Exit(1)
        }
    }

    // Process months flag
    if *monthsFlag != 0 {
        if *monthsFlag == 1 || *monthsFlag == 3 || *monthsFlag == 6 || *monthsFlag == 12 {
            cfg.NumMonths = *monthsFlag
        } else {
            fmt.Fprintf(os.Stderr, "Error: Invalid months value %d. Must be 1, 3, 6, or 12.\n", *monthsFlag)
            flag.Usage()
            os.Exit(1)
        }
    }


    // Process months flag
    if *columnsFlag != 0 {
        if *columnsFlag == 1 || *columnsFlag == 2 || *columnsFlag == 3|| *columnsFlag == 4 || *columnsFlag == 6 || *columnsFlag == 12 {
            cfg.NumColumns = *columnsFlag
        } else {
            fmt.Fprintf(os.Stderr, "Error: Invalid columns value %d. Must be 1, 2, 3, 4, 6, or 12.\n", *columnsFlag)
            flag.Usage()
            os.Exit(1)
        }
    }


    // Process display flag
    switch *displayFlag {
    case DisplayCalendar, DisplayEvents, DisplayBoth:
        cfg.DisplayMode = *displayFlag
    default:
        fmt.Fprintf(os.Stderr, "Error: Invalid display value '%s'. Must be 'calendar', 'events', or 'both'.\n", *displayFlag)
        flag.Usage()
        os.Exit(1)
    }

    // Determine the actual starting month and year for display
    displayMonth, displayYear := getDisplayMonthYear(cfg)

    // Load events for the range of years that will be displayed.
    // If displaying multiple months, especially 12, it might span two calendar years.
    yearsToLoad := []int{displayYear}
    if cfg.NumMonths > 1 && (int(displayMonth)+cfg.NumMonths-1 > 12) {
        // If the range of months extends into the next year, add the next year to load events for it.
        yearsToLoad = append(yearsToLoad, displayYear+1)
    }

    var allEvents []Event
    for _, year := range yearsToLoad {
        eventsForYear, err := LoadEvents(cfg.EventsFile, year)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Warning: Could not load events for year %d: %v\n", year, err)
            // Continue, don't exit, just print a warning
        }
        allEvents = append(allEvents, eventsForYear...)
    }

    // Print based on DisplayMode
    if cfg.DisplayMode == DisplayCalendar || cfg.DisplayMode == DisplayBoth {
        PrintCalendar(cfg, displayMonth, displayYear, allEvents)
    }
    if cfg.DisplayMode == DisplayEvents || cfg.DisplayMode == DisplayBoth {
        PrintEventList(cfg, displayMonth, displayYear, allEvents)
    }
}

