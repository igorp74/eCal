# eCal

Command line calendar with events

Inspired by win Calendar application by Horst Schaeffer https://www.horstmuc.de/wrem.htm#calendar

## Usage
 `calendar` [options]

**Options**:
|Flag|Description|Default|
|:--|:--|--:|
| `-c string` | Number of columns to display (1, 2, 3, 4, 6, or 12) | 1 |
| `-d string` | What to display: 'calendar', 'events', or 'both' | "both" |
| `-f string` | Path to the events file. | `"events.txt"` |
| `-m int` | Month for the calendar (1-12) | current month |
| `-mn int` | Number of months to display (1, 3, 6, or 12). (default 1) |
| `-monday` | Set Monday as the first day of the week. | `true` |
| `-w int` | Week number for the calendar (1-53). If used with `-y`, overrides `-m`| |
| `-wk` | Show week numbers. | `true` |
| `-y int` | Year for the calendar. Also used with `-w`. | current year |


# Documentation
* [⚙️ Build](https://github.com/igorp74/eCal/wiki/%E2%9A%99%EF%B8%8F-Build)
* [🎬 Examples](https://github.com/igorp74/eCal/wiki/%F0%9F%8E%AC-Examples)
