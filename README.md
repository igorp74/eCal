# eCal

Command line calendar with events

Inspired by win Calendar application by Horst Schaeffer https://www.horstmuc.de/wrem.htm#calendar

## Usage
 `calendar` [options]

**Options**:
|Flag|Description|Default|
|:--|:--|--:|
| `-display string` | What to display: 'calendar', 'events', or 'both' | "both" |
| `-events string` | Path to the events file. | `"events.txt"` |
| `-mondayFirst` | Set Monday as the first day of the week. | `true` |
| `-month int` | Month for the calendar (1-12) | current month |
| `-months int` | Number of months to display (1, 3, 6, or 12). (default 1) |
| `-week int` | Week number for the calendar (1-53). If used with `-year`, overrides `-month`| |
| `-weeknumbers` | Show week numbers. | `true` |
| `-year int` | Year for the calendar. Also used with `-week`. | current year |
