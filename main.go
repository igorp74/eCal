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
        TargetTime:  currentTime,  // Reference time for age/countdown
    }

    // Command-line flags
    yearFlag  := flag.Int("year", 0, "Year for the calendar (default: current year). Also used with -week.")
    monthFlag := flag.Int("month",0, "Month for the calendar (1-12) (default: current month).")
    weekFlag  := flag.Int("week", 0, "Week number for the calendar (1-53). If used with -year, overrides -month.")

    flag.BoolVar(&cfg.MondayFirst,  "mondayFirst", cfg.MondayFirst, "Set Monday as the first day of the week.")
    flag.StringVar(&cfg.EventsFile, "events", cfg.EventsFile, "Path to the events file.")
    flag.BoolVar(&cfg.ShowWeekNum,  "weeknumbers", cfg.ShowWeekNum, "Show week numbers.")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        fmt.Fprintf(os.Stderr, " %s [options]\n\nOptions:\n", os.Args[0])
        flag.PrintDefaults()
        fmt.Fprintf(os.Stderr, "\nExamples:\n")
        fmt.Fprintf(os.Stderr, "  %s -year 2024 -month 12\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s -year 2024 -week 50\n", os.Args[0])
        fmt.Fprintf(os.Stderr, "  %s -events my_holidays.txt -mondayFirst\n", os.Args[0])
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


    // Determine the actual month and year to display
    displayMonth, displayYear := getDisplayMonthYear(cfg)

    // Load events for the determined display year
    allEvents, err := LoadEvents(cfg.EventsFile, displayYear)
    if err != nil {
        // LoadEvents already prints info/warnings, but a fatal error might stop here
        fmt.Fprintf(os.Stderr, "Critical error loading events: %v\n", err)
        // os.Exit(1); // Or continue with no events
    }

    // Print the calendar
    PrintCalendar(cfg, displayMonth, displayYear, allEvents)
}
