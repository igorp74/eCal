# eCal

Command line calendar with events

Inspired by win Calendar application by Horst Schaeffer https://www.horstmuc.de/wrem.htm#calendar

## Usage
 `calendar` [options]

**Options**:
|Flag|Description|
|:--|:--|
| `-events string` | Path to the events file. (default "events.txt") |
| `-mondayFirst` | Set Monday as the first day of the week. (default true) |
| `-month int` | Month for the calendar (1-12) (default: current month).|
| `-week int` | Week number for the calendar (1-53). If used with -year, overrides -month.|
| `-weeknumbers` | Show week numbers. (default true)|
| `-year int` | Year for the calendar (default: current year). Also used with -week.|
  
**Examples**:

  `calendar -year 2024 -month 12`
  
  `calendar -year 2024 -week 50`
  
  `calendar -events my_holidays.txt -mondayFirst`

**In action**

![ecal](https://github.com/user-attachments/assets/27cad616-feed-4585-ab2d-516855b9cd9c)
