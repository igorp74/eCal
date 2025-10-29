package main

import (
    "strings" // Import the strings package
    "time"
)

// ANSI color escape codes
const (
    style_reset     = "\033[0m"
    style_bold      = "\033[1m"
    style_italic    = "\033[3m"
    style_underline = "\033[4m"

    fg_black   = "\033[30m"
    fg_red     = "\033[31m"
    fg_green   = "\033[32m"
    fg_yellow  = "\033[33m"
    fg_blue    = "\033[34m"
    fg_magenta = "\033[35m"
    fg_cyan    = "\033[36m"
    fg_white   = "\033[37m"

    bg_black   = "\033[40m"
    bg_red     = "\033[41m"
    bg_green   = "\033[42m"
    bg_yellow  = "\033[43m"
    bg_blue    = "\033[44m"
    bg_magenta = "\033[45m"
    bg_cyan    = "\033[46m"
    bg_white   = "\033[47m"
)

// DisplayMode constants
const (
    DisplayBoth     = "both"
    DisplayCalendar = "calendar"
    DisplayEvents   = "events"
)

// Config holds the application's runtime configuration
type Config struct {
    Year        int
    Month       time.Month
    Week        int // If Week > 0, it's used with Year to determine month
    MondayFirst bool
    EventsFile  string
    ShowWeekNum bool
    TargetTime  time.Time // Current time for age/countdown calculations
    NumMonths   int       // New field: Number of months to display (1, 3, 6, 12)
    DisplayMode string    // New field: "calendar", "events", or "both"
}

// Event represents a calendar event
type Event struct {
    Date             time.Time // Actual date of the event for the target year/month
    OriginalDateStr  string    // Original date string from the event file
    Description      string
    Type             string    // e.g., "birthday", "ie", "us"
    IsAnnual         bool      // True if the event occurs annually without a fixed year in its rule
    IsAnniversary    bool      // True if the event type is "anniversary" and a birth year is known
    AnniDate         time.Time // Full birth date (YYYY-MM-DD) if available
    RecurrenceRule   string    // Stores the rule string like "E+1", "MM/DOW#N" for reference
    SpecificYearRule bool      // True if the event rule itself specified a year (e.g., MM/DD/YYYY, MM/DD?YYYY)
    DisplayColor     string    // ANSI foreground color code for highlighting this event type
    DisplayBgColor   string    // ANSI background color code for highlighting this event type
    Emoji            string    // New field: Specific emoji for this event, if provided
}

// getDefaultEmoji returns a default emoji for a given event type.
// This is used if no specific emoji is provided in the events.ini for an event.
func getDefaultEmoji(tag string) string {
    switch tag {
    case "global":
        return "🌐"
    case "anniversary":
        return "📌"
    case "birthday":
        return "🎂"
    case "hr":
        return "🇭🇷"
    case "ie":
        return "🇮🇪"
    case "us":
        return "🇺🇸"
    case "holiday":
        return "🏖️"
    case "church":
        return "✝️"
    case "fun":
        return "🎉"
    default:
        return "📅"
    }
}

// GetFgColorCode returns the ANSI foreground color code for a given color name.
// It defaults to fg_green if the color name is not recognized.
func GetFgColorCode(colorName string) string {
    switch strings.ToLower(colorName) {
    case "black":
        return fg_black
    case "red":
        return fg_red
    case "green":
        return fg_green
    case "yellow":
        return fg_yellow
    case "blue":
        return fg_blue
    case "magenta":
        return fg_magenta
    case "cyan":
        return fg_cyan
    case "white":
        return fg_white
    default:
        return fg_white // Default to green if color is not specified or recognized
    }
}

// GetBgColorCode returns the ANSI background color code for a given color name.
// It returns an empty string if the color name is not recognized, meaning no background color.
func GetBgColorCode(colorName string) string {
    switch strings.ToLower(colorName) {
    case "black":
        return bg_black
    case "red":
        return bg_red
    case "green":
        return bg_green
    case "yellow":
        return bg_yellow
    case "blue":
        return bg_blue
    case "magenta":
        return bg_magenta
    case "cyan":
        return bg_cyan
    case "white":
        return bg_white
    default:
        return "" // Default to no background color if not specified or recognized
    }
}

