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
| `-months int` | Number of months to display (1, 3, 6, or 12). (default 1) |
| `-week int` | Week number for the calendar (1-53). If used with `-year`, overrides `-month`| |
| `-weeknumbers` | Show week numbers. | `true` |
| `-year int` | Year for the calendar. Also used with `-week`. | current year |
  
**Examples**:

  `calendar -year 2024 -month 12`
  
  `calendar -year 2024 -week 50`
  
  `calendar -events my_holidays.txt -mondayFirst`

**In action**

A single month

![image](https://github.com/user-attachments/assets/fe625880-e807-4eb2-a46a-4281acc05dcc)

3 months view

![image](https://github.com/user-attachments/assets/9eb1b220-d324-4cfa-bef3-549742362bd9)

6 months view

![image](https://github.com/user-attachments/assets/7672f289-f924-41e7-a8f3-9012e180b2b5)
