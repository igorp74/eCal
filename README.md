# eCal

Command line calendar with events

Inspired by win Calendar application by Horst Schaeffer https://www.horstmuc.de/wrem.htm#calendar

## Usage
 `calendar` [options]

**Options**:
|Flag|Description|Default|
|:--|:--|--:|
| `-events string` | Path to the events file. | `"events.txt"` |
| `-mondayFirst` | Set Monday as the first day of the week. | `true` |
| `-month int` | Month for the calendar (1-12) | current month |
| `-week int` | Week number for the calendar (1-53). If used with `-year`, overrides `-month`| |
| `-weeknumbers` | Show week numbers. | `true` |
| `-year int` | Year for the calendar. Also used with `-week`. | current year |
  
**Examples**:

  `calendar -year 2024 -month 12`
  
  `calendar -year 2024 -week 50`
  
  `calendar -events my_holidays.txt -mondayFirst`

**In action**

![image](https://github.com/user-attachments/assets/2ec26067-4f1c-481b-bafd-bb93f168968d)

