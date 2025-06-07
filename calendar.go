package main

import (
    "fmt"
    "os"
    "time"
)

// getDisplayMonthYear determines the target month and year for the calendar display
// based on the configuration (either month/year or year/week).
func getDisplayMonthYear(cfg Config) (time.Month, int) {
    if cfg.Week > 0 && cfg.Year > 0 {
        // Calculate month/year from year/week input
        firstDayOfWeek, err := GetFirstDayOfISOWeek(cfg.Year, cfg.Week)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error determining date from year/week: %v. Falling back to current month/year.\n", err)
            now := time.Now()
            return now.Month(), now.Year()
        }
        return firstDayOfWeek.Month(), firstDayOfWeek.Year()
    }
    // Default to configured month and year (which might be current if not specified)
    return cfg.Month, cfg.Year
}
