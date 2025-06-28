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
	}

	// Command-line flags
	yearFlag := flag.Int("year", 0, "Year for the calendar (default: current year). Also used with -week.")
	monthFlag := flag.Int("month", 0, "Month for the calendar (1-12) (default: current month).")
	weekFlag := flag.Int("week", 0, "Week number for the calendar (1-53). If used with -year, overrides -month.")
	monthsFlag := flag.Int("months", 1, "Number of months to display (1, 3, 6, or 12).") // New flag

	flag.BoolVar(&cfg.MondayFirst, "mondayFirst", cfg.MondayFirst, "Set Monday as the first day of the week.")
	flag.StringVar(&cfg.EventsFile, "events", cfg.EventsFile, "Path to the events file.")
	flag.BoolVar(&cfg.ShowWeekNum, "weeknumbers", cfg.ShowWeekNum, "Show week numbers.")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, " %s [options]\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -year 2024 -month 12\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -year 2024 -week 50\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -events my_holidays.txt -mondayFirst\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -month 7 -year 2025 -months 3\n", os.Args[0]) // New example
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

	// Process new months flag
	if *monthsFlag != 0 {
		if *monthsFlag == 1 || *monthsFlag == 3 || *monthsFlag == 6 || *monthsFlag == 12 {
			cfg.NumMonths = *monthsFlag
		} else {
			fmt.Fprintf(os.Stderr, "Error: Invalid months value %d. Must be 1, 3, 6, or 12.\n", *monthsFlag)
			flag.Usage()
			os.Exit(1)
		}
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

	// Print the calendar(s)
	PrintMultiMonthCalendar(cfg, displayMonth, displayYear, allEvents)
}
