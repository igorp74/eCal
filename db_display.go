package main

import (
    "database/sql"
    "fmt"
    "log"
    "strconv"
    "strings"
    "time"
)

// Display formats
const (
    DisplayFull = iota
    DisplayCondensed
    DisplayMinimal

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

// ListTasks fetches and displays tasks based on filters and sorting.
// Added endBefore and endAfter parameters for filtering by end date.
// Added taskIDs for filtering by specific task IDs, and searchText for title/description/note search.
func ListTasks(tm *TodoManager, projectFilter, contextFilter, tagFilter, statusFilter, startBefore, startAfter, dueBefore, dueAfter, endBefore, endAfter, sortBy, order string, format int, displayNotes string, taskIDs []int64, searchText string) {
    // Base query to select task details.
    // LEFT JOIN is used for projects, and now for task_notes to allow searching within notes
    // without requiring every task to have notes.
    query := `
        SELECT
            t.id, t.title, t.description, p.name, t.start_date, t.due_date, t.end_date, t.status,
            t.recurrence, t.recurrence_interval, t.start_waiting_date, t.end_waiting_date, t.original_task_id,
            COALESCE(tm.created_at, t.start_date) as created_at_ts -- Get created_at from task_metadata or fallback to start_date
        FROM tasks t
        LEFT JOIN projects p ON t.project_id = p.id
        LEFT JOIN task_metadata tm ON t.id = tm.task_id -- Join with new metadata table
    `
    args := []any{}
    whereClauses := []string{"1=1"} // Start with a true condition to simplify AND logic

    // Filter by specific task IDs
    if len(taskIDs) > 0 {
        placeholders := make([]string, len(taskIDs))
        for i := range taskIDs {
            placeholders[i] = "?"
            args = append(args, taskIDs[i])
        }
        whereClauses = append(whereClauses, fmt.Sprintf("t.id IN (%s)", strings.Join(placeholders, ",")))
    }

    // Project filter
    if projectFilter != "" {
        whereClauses = append(whereClauses, "p.name = ?")
        args = append(args, projectFilter)
    }

    // Status filter
    if statusFilter != "" && statusFilter != "all" {
        whereClauses = append(whereClauses, "t.status = ?")
        args = append(args, statusFilter)
    }

    // Search text filter in title, description, and notes
    if searchText != "" {
        searchPattern := "%" + searchText + "%"
        // Use a subquery with EXISTS to check for matching notes
        whereClauses = append(whereClauses, `(
            t.title LIKE ? OR t.description LIKE ?
            OR EXISTS (SELECT 1 FROM task_notes tn WHERE tn.task_id = t.id AND tn.description LIKE ?)
        )`)
        args = append(args, searchPattern, searchPattern, searchPattern)
    }

    // Date filters - parse with local timezone and then convert to UTC for query
    if startBefore != "" {
        parsed, err := ParseDateTime(startBefore, time.Local)
        if err != nil {
            log.Printf("Warning: Invalid start-before date format: %v", err)
        } else {
            whereClauses = append(whereClauses, "t.start_date <= ?")
            sqlParsed, _ := parsed.Value() // Get sql.NullTime (already UTC)
            args = append(args, sqlParsed)
        }
    }
    if startAfter != "" {
        parsed, err := ParseDateTime(startAfter, time.Local)
        if err != nil {
            log.Printf("Warning: Invalid start-after date format: %v", err)
        } else {
            whereClauses = append(whereClauses, "t.start_date >= ?")
            sqlParsed, _ := parsed.Value() // Get sql.NullTime (already UTC)
            args = append(args, sqlParsed)
        }
    }
    if dueBefore != "" {
        parsed, err := ParseDateTime(dueBefore, time.Local)
        if err != nil {
            log.Printf("Warning: Invalid due-before date format: %v", err)
        } else {
            whereClauses = append(whereClauses, "t.due_date <= ?")
            sqlParsed, _ := parsed.Value() // Get sql.NullTime (already UTC)
            args = append(args, sqlParsed)
        }
    }
    if dueAfter != "" {
        parsed, err := ParseDateTime(dueAfter, time.Local)
        if err != nil {
            log.Printf("Warning: Invalid due-after date format: %v", err)
        } else {
            whereClauses = append(whereClauses, "t.due_date >= ?")
            sqlParsed, _ := parsed.Value() // Get sql.NullTime (already UTC)
            args = append(args, sqlParsed)
        }
    }
    // New: End Date filters
    if endBefore != "" {
        parsed, err := ParseDateTime(endBefore, time.Local)
        if err != nil {
            log.Printf("Warning: Invalid end-before date format: %v", err)
        } else {
            whereClauses = append(whereClauses, "t.end_date <= ?")
            sqlParsed, _ := parsed.Value() // Get sql.NullTime (already UTC)
            args = append(args, sqlParsed)
        }
    }
    if endAfter != "" {
        parsed, err := ParseDateTime(endAfter, time.Local)
        if err != nil {
            log.Printf("Warning: Invalid end-after date format: %v", err)
        } else {
            whereClauses = append(whereClauses, "t.end_date >= ?")
            sqlParsed, _ := parsed.Value() // Get sql.NullTime (already UTC)
            args = append(args, sqlParsed)
        }
    }

    // Context and Tag filters (require JOINs and GROUP BY or EXISTS subqueries)
    if contextFilter != "" {
        whereClauses = append(whereClauses, `EXISTS (SELECT 1 FROM task_contexts tc JOIN contexts c ON tc.context_id = c.id WHERE tc.task_id = t.id AND c.name = ?)`)
        args = append(args, contextFilter)
    }
    if tagFilter != "" {
        whereClauses = append(whereClauses, `EXISTS (SELECT 1 FROM task_tags tt JOIN tags tg ON tt.tag_id = tg.id WHERE tt.task_id = t.id AND tg.name = ?)`)
        args = append(args, tagFilter)
    }

    // Combine all WHERE clauses
    query += " WHERE " + strings.Join(whereClauses, " AND ")

    // Order by
    orderByMap := map[string]string{
        "id":         "t.id",
        "title":      "t.title",
        "start_date": "t.start_date",
        "due_date":   "t.due_date",
        "status":     "t.status",
        "project":    "p.name",
        "end_date":   "t.end_date",
    }
    actualSortBy := orderByMap[sortBy]
    if actualSortBy == "" {
        actualSortBy = orderByMap["due_date"] // Default
    }
    if order != "asc" && order != "desc" {
        order = "asc" // Default
    }
    query += fmt.Sprintf(" ORDER BY %s %s", actualSortBy, order)

    rows, err := tm.db.Query(query, args...)
    if err != nil {
        log.Fatalf("Error querying tasks: %v", err)
    }
    defer rows.Close()

    // Load working hours and holidays once for all calculations
    workingHours, err := tm.GetWorkingHours()
    if err != nil {
        log.Fatalf("Error loading working hours: %v", err)
    }
    holidaysList, err := tm.GetHolidays()
    if err != nil {
        log.Fatalf("Error loading holidays: %v", err)
    }
    holidaysMap := make(map[string]Holiday)
    for _, h := range holidaysList {
        holidaysMap[h.Date.Time.Format("2006-01-02")] = h
    }

    // Print header based on format
    switch format {
    case DisplayMinimal:
        fmt.Println("----------------------------------------------------------------------------------------------------------------")
        fmt.Printf("%-5s    %-23s  %-20s %-80s\n", "ID", "Created", "Project", "Title")
        fmt.Println("----------------------------------------------------------------------------------------------------------------")
    }

    for rows.Next() {
        var task Task
        var project_name sql.NullString
        var desc, recurrence sql.NullString
        var recurrenceInterval sql.NullInt64
        var startDate, dueDate, endDate, startWaitingDate, endWaitingDate sql.NullTime
        var originalTaskID sql.NullInt64
        var createdAtStr sql.NullString // Changed to scan created_at into a string

        err := rows.Scan(&task.ID, &task.Title, &desc, &project_name, &startDate, &dueDate, &endDate, &task.Status,
            &recurrence, &recurrenceInterval, &startWaitingDate, &endWaitingDate, &originalTaskID, &createdAtStr) // Scan into string
        if err != nil {
            log.Printf("Error scanning task: %v", err)
            continue
        }

        task.Description = desc
        task.ProjectName = project_name
        task.StartDate = NullableTime{Time: startDate.Time, Valid: startDate.Valid}
        task.DueDate = NullableTime{Time: dueDate.Time, Valid: dueDate.Valid}
        task.EndDate = NullableTime{Time: endDate.Time, Valid: endDate.Valid}
        task.Recurrence = recurrence
        task.RecurrenceInterval = recurrenceInterval
        task.StartWaitingDate = NullableTime{Time: startWaitingDate.Time, Valid: startWaitingDate.Valid}
        task.EndWaitingDate = NullableTime{Time: endWaitingDate.Time, Valid: endWaitingDate.Valid}
        task.OriginalTaskID = originalTaskID

        var createdAt NullableTime
        if createdAtStr.Valid {
            // Parse the string into NullableTime
            parsedCreatedAt, parseErr := ParseDateTime(createdAtStr.String, time.UTC) // Assume UTC for DB-stored timestamps
            if parseErr != nil {
                log.Printf("Warning: Could not parse created_at timestamp '%s' for task %d: %v", createdAtStr.String, task.ID, parseErr)
            } else {
                createdAt = parsedCreatedAt
            }
        }

        // Fetch contexts and tags
        task.Contexts = tm.GetTaskNames(int64(task.ID), "task_contexts", "contexts")
        task.Tags = tm.GetTaskNames(int64(task.ID), "task_tags", "tags")

        // Fetch notes based on displayNotes parameter
        if displayNotes != "none" {
            allNotes := tm.GetNotesForTask(task.ID)
            if displayNotes == "all" {
                task.Notes = allNotes
            } else {
                numNotes, err := strconv.Atoi(displayNotes)
                if err == nil && numNotes > 0 {
                    if numNotes > len(allNotes) {
                        task.Notes = allNotes
                    } else {
                        task.Notes = allNotes[len(allNotes)-numNotes:]
                    }
                }
            }
        }

        // Calculate Duration and Working Hours Duration
        totalDurationStr := "N/A"
        workingDurationStr := "N/A"
        waitingDurationStr := "N/A"
        waitingWorkingDurationStr := "N/A"
        timeToDueStr := ""

        if task.StartDate.Valid {
            if task.Status == "completed" && task.EndDate.Valid {
                totalDuration := CalculateCalendarDuration(task)
                totalDurationStr = FormatDuration(totalDuration)

                workingDuration := tm.CalculateWorkingDuration(task.StartDate, task.EndDate, workingHours, holidaysMap)
                workingDurationStr = FormatWorkingHoursDisplay(workingDuration)
            } else if task.Status != "completed" {
                tempTask := task
                tempTask.EndDate = NullableTime{Time: time.Now().UTC(), Valid: true}
                totalDuration := CalculateCalendarDuration(tempTask)
                totalDurationStr = FormatDuration(totalDuration)

                workingDuration := tm.CalculateWorkingDuration(task.StartDate, NullableTime{Time: time.Now().UTC(), Valid: true}, workingHours, holidaysMap)
                workingDurationStr = FormatWorkingHoursDisplay(workingDuration)
            }
        }

        // Calculate time to/after due date
        if task.DueDate.Valid {
            diffDuration, isOverdue := CalculateTimeDifference(task.DueDate)
            if isOverdue {
                timeToDueStr = fmt.Sprintf(" (%s%s%s overdue)", fg_red, FormatDuration(diffDuration), style_reset)
            } else {
                timeToDueStr = fmt.Sprintf(" (%s%s%s remaining)", fg_cyan, FormatDuration(diffDuration), style_reset)
            }
        }

        // Calculate waiting duration (calendar time)
        waitingDuration := CalculateWaitingDuration(task)
        waitingDurationStr = FormatDuration(waitingDuration)

        // Calculate working hours within the waiting period
        if task.StartWaitingDate.Valid && task.EndWaitingDate.Valid {
            waitingWorkingDuration := tm.CalculateWorkingDuration(task.StartWaitingDate, task.EndWaitingDate, workingHours, holidaysMap)
            waitingWorkingDurationStr = FormatWorkingHoursDisplay(waitingWorkingDuration)
        }

        switch format {
        case DisplayFull:
            var sb strings.Builder
            sb.WriteString(fmt.Sprintf("\n%s%-5d%s", fg_red, task.ID, style_reset))

            titleParts := []string{style_bold + task.Title + style_reset}

            status_str := ""
            switch task.Status {
            case "pending":
                status_str = style_bold + fg_yellow + "pending" + style_reset
            case "completed":
                status_str = style_bold + fg_green + "completed" + style_reset
            case "cancelled":
                status_str = style_bold + fg_red + "canceled" + style_reset
            case "waiting":
                status_str = style_bold + fg_blue + "waiting" + style_reset
            }
            titleParts = append(titleParts, status_str)

            if task.Recurrence.Valid {
                interval := ""
                if task.RecurrenceInterval.Valid {
                    interval = fmt.Sprintf(" every %d", task.RecurrenceInterval.Int64)
                }
                titleParts = append(titleParts, "🔄 "+fg_blue+task.Recurrence.String+interval+style_reset)

            }
            if task.DueDate.Valid {
                titleParts = append(titleParts, style_bold+timeToDueStr+style_reset)
            }

            sb.WriteString(fmt.Sprintf(" %s\n", strings.Join(titleParts, " | ")))


            if task.Description.Valid && task.Description.String != "" {
                sb.WriteString(fmt.Sprintf("      📜 %s%s%s%s\n", style_italic, fg_yellow, task.Description.String, style_reset))
            }



            projectParts := []string{}

            if len(task.ProjectName.String) > 0 {
                projectParts = append(projectParts, "📌 Project: "+fg_green+task.ProjectName.String+style_reset)
            }
            if len(task.Tags) > 0 {
                projectParts = append(projectParts, "🏷️ Tags: "+fg_blue+strings.Join(task.Tags, ", ")+style_reset)
            }
            if len(task.Contexts) > 0 {
                projectParts = append(projectParts, "🔖 Context: "+fg_magenta+strings.Join(task.Contexts, ", ")+style_reset)
            }
            if len(projectParts) > 0 {
                sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(projectParts, " | ")))
            }



            timeParts := []string{}

            // if createdAt.Valid {
            //     timeParts = append(timeParts, "✨ Added: "+FormatDisplayDateTime(createdAt))
            // }
            if task.DueDate.Valid {
                timeParts = append(timeParts, "⏰ Due:   "+FormatDisplayDateTime(task.DueDate))
            }

            if len(timeParts) > 0 {
                sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(timeParts, " | ")))
            }


            // if task.Recurrence.Valid {
            //     interval := ""
            //     if task.RecurrenceInterval.Valid {
            //         interval = fmt.Sprintf(" every %d", task.RecurrenceInterval.Int64)
            //     }
            //     sb.WriteString(fmt.Sprintf("      🔄 Recurrence: %s%s%s%s\n", fg_blue, task.Recurrence.String, interval, style_reset))
            // }


            // if createdAt.Valid {
            //     sb.WriteString(fmt.Sprintf("      ✨ Created: %s\n", FormatDisplayDateTime(createdAt)))
            // }

            dateParts := []string{}

            if task.StartDate.Valid {
                dateParts = append(dateParts, "🚀 Start: "+FormatDisplayDateTime(task.StartDate))
            }
            if task.EndDate.Valid {
                dateParts = append(dateParts, "🏁 End: "+FormatDisplayDateTime(task.EndDate))
            }
            if len(totalDurationStr) > 0 && totalDurationStr != "N/A" {
                dateParts = append(dateParts, "(" + fg_green + totalDurationStr + style_reset + ")")
            }

            if len(dateParts) > 0 {
                sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(dateParts, " | ")))
            }


            waitingParts := []string{}

            if task.StartWaitingDate.Valid {
                waitingParts = append(waitingParts, "⏸️ Pause: "+FormatDisplayDateTime(task.StartWaitingDate))
            }
            if task.EndWaitingDate.Valid {
                waitingParts = append(waitingParts, "▶️ End: "+FormatDisplayDateTime(task.EndWaitingDate))
            }
            if waitingDurationStr != "0s" && waitingDurationStr != "N/A" { // Only add if there's a non-zero waiting calendar duration
                waitingParts = append(waitingParts, "(" + fg_red + waitingDurationStr + style_reset + ")")
            }

            if len(waitingParts) > 0 {
                sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(waitingParts, " | ")))
            }



            durationParts := []string{}
            if workingDurationStr != "0s" && workingDurationStr != "N/A" {
                durationParts = append(durationParts, "⏳ Working: "+style_bold+fg_green+" "+workingDurationStr+" "+style_reset)
            }
            if waitingWorkingDurationStr != "0s" && waitingWorkingDurationStr != "N/A"  { // Only add if there's a non-zero waiting working duration
                durationParts = append(durationParts, "🚧 Waiting: "+fg_red+waitingWorkingDurationStr+style_reset)
            }

            if len(durationParts) > 0 {
                sb.WriteString(fmt.Sprintf("      %s\n", strings.Join(durationParts, " | ")))
            }



            // Display Notes
            if len(task.Notes) > 0 {
                sb.WriteString(fmt.Sprintf("      📝 %sNotes:%s\n", style_bold, style_reset))
                // Iterate backwards to display newest (largest ID) first
                for j := len(task.Notes) - 1; j >= 0; j-- {
                    note := task.Notes[j]
                    if note.Timestamp.Valid && note.Description.Valid {
                        // Changed to display actual note.ID instead of a calculated display ID
                        sb.WriteString(fmt.Sprintf("         %-5d %s%s%s%s: %s%s%s\n", note.ID, style_italic, fg_green, FormatDisplayDateTime(note.Timestamp), style_reset, fg_yellow, note.Description.String, style_reset))
                    }
                }
            }

            fmt.Printf("%s", sb.String())

        case DisplayCondensed:

            var sb strings.Builder
            sb.WriteString(fmt.Sprintf("\n%s%-5d%s", fg_red, task.ID, style_reset))

            titleParts := []string{}

            status_str := ""
            switch task.Status {
            case "pending":
                status_str = "🚀"
            case "completed":
                status_str = "✅"
            case "cancelled":
                status_str = "❌"
            case "waiting":
                status_str = "⏸️"
            }
            titleParts = append(titleParts, status_str)
            titleParts = append(titleParts, style_bold+task.Title+style_reset)
            if task.Recurrence.Valid {
                interval := ""
                if task.RecurrenceInterval.Valid {
                    interval = fmt.Sprintf(" every %d", task.RecurrenceInterval.Int64)
                }
                titleParts = append(titleParts, "  🔄 "+fg_blue+task.Recurrence.String+interval+style_reset)

            }
            if task.DueDate.Valid {
                titleParts = append(titleParts, style_bold+timeToDueStr+style_reset)
            }

            sb.WriteString(fmt.Sprintf(" %s\n", strings.Join(titleParts, " ")))


            if task.Description.Valid && task.Description.String != "" {
                sb.WriteString(fmt.Sprintf("         %s%s%s%s\n", style_italic, fg_yellow, task.Description.String, style_reset))
            }


            // if createdAt.Valid {
            //     sb.WriteString(fmt.Sprintf("         ✨ Created: %s\n", FormatDisplayDateTime(createdAt)))
            // }

            // Add due date and time to due
            // if task.DueDate.Valid {
            //     sb.WriteString(fmt.Sprintf("         %s%s%s\n", FormatDisplayDateTime(task.DueDate), timeToDueStr, style_reset))
            // }


            projectParts := []string{}

            if len(task.ProjectName.String) > 0 {
                projectParts = append(projectParts, fg_green+task.ProjectName.String+style_reset)
            }
            if len(task.Tags) > 0 {
                projectParts = append(projectParts, fg_blue+strings.Join(task.Tags, ", ")+style_reset)
            }
            if len(task.Contexts) > 0 {
                projectParts = append(projectParts, fg_magenta+strings.Join(task.Contexts, ", ")+style_reset)
            }
            if len(projectParts) > 0 {
                sb.WriteString(fmt.Sprintf("         %s\n", strings.Join(projectParts, " | ")))
            }



            // durationParts := []string{}
            // if workingDurationStr != "0s" && workingDurationStr != "N/A" {
            //     durationParts = append(durationParts, style_bold+fg_green+workingDurationStr+style_reset)
            // }
            // if waitingWorkingDurationStr != "0s" && waitingWorkingDurationStr != "N/A" { // Only add if there's a non-zero waiting working duration
            //     durationParts = append(durationParts, fg_red+waitingWorkingDurationStr+style_reset)
            // }
            // if len(totalDurationStr) > 0 && totalDurationStr != "N/A" {
            //     durationParts = append(durationParts, "(" + fg_green + totalDurationStr + style_reset + ")")
            // }
            // if len(durationParts) > 0 {
            //     sb.WriteString(fmt.Sprintf("         %s\n", strings.Join(durationParts, "  ")))
            // }



            // Display Notes
            if len(task.Notes) > 0 {
                sb.WriteString("         📝-----------------------------\n")
                // Iterate backwards to display newest (largest ID) first
                for j := len(task.Notes) - 1; j >= 0; j-- {
                    note := task.Notes[j]
                    if note.Timestamp.Valid && note.Description.Valid {
                        // Changed to display actual note.ID instead of a calculated display ID
                        sb.WriteString(fmt.Sprintf("         %-5d %s%s%s%s: %s%s%s\n", note.ID, style_italic, fg_green, FormatDisplayDateTime(note.Timestamp), style_reset, fg_yellow, note.Description.String, style_reset))
                    }
                }
            }

            fmt.Printf("%s", sb.String())

        case DisplayMinimal:
            status_str := ""
            switch task.Status {
            case "pending":
                status_str = "🚀"
            case "completed":
                status_str = "✅"
            case "cancelled":
                status_str = "❌"
            case "waiting":
                status_str = "⏸️"
            }

            fmt.Printf("%-5d%s  %-23s  %s%-20s%s %s%-80s%s\n",
                task.ID,
                status_str,
                FormatDisplayDateTime(createdAt),
                fg_green, task.ProjectName.String, style_reset,
                style_bold, task.Title, style_reset)
        }
    }
    fmt.Println("----------------------------------------------------------------------------------------------------------------")
}

// ListHolidays lists all configured holidays.
// It now accepts *TodoManager.
func ListHolidays(tm *TodoManager) {
    holidays, err := tm.GetHolidays() // Get as slice
    if err != nil {
        log.Fatalf("Error listing holidays: %v", err)
    }

    fmt.Println("--- Holidays ---")
    fmt.Println("  ID    Date        Name") // New header with ID
    fmt.Println("------------------------------")
    if len(holidays) == 0 {
        fmt.Println("No holidays configured.")
        return
    }
    for _, h := range holidays { // Iterate over slice
        // Holidays are stored asYYYY-MM-DD strings, no time component.
        // Display them directly.
        fmt.Printf("  %-5d %-10s %s\n", h.ID, h.Date.Time.Format("2006-01-02"), h.Name) // Print ID and formatted date
    }
}

// ListWorkingHours lists all configured working hours.
// It now accepts *TodoManager.
func ListWorkingHours(tm *TodoManager) {
    // Query all columns relevant to working hours, including minutes and break duration.
    rows, err := tm.db.Query("SELECT day_of_week, start_hour, start_minute, end_hour, end_minute, break_minutes FROM working_hours ORDER BY day_of_week ASC")
    if err != nil {
        log.Fatalf("Error listing working hours: %v", err)
    }
    defer rows.Close()

    fmt.Println("--- Working Hours ---")
    found := false
    dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
    for rows.Next() {
        found = true
        var day, startHour, startMinute, endHour, endMinute, breakMinutes int
        // Scan all fetched columns into respective variables.
        if err := rows.Scan(&day, &startHour, &startMinute, &endHour, &endMinute, &breakMinutes); err != nil {
            log.Printf("Error scanning working hours: %v", err)
            continue
        }
        if day >= 0 && day < len(dayNames) {
            // Print working hours including minutes and break duration.
            fmt.Printf("  %-10s %02d:%02d - %02d:%02d (Break: %d minutes)\n", dayNames[day], startHour, startMinute, endHour, endMinute, breakMinutes)
        } else {
            fmt.Printf("  Day %d: %02d:%02d - %02d:%02d (Break: %d minutes) (Invalid Day Index)\n", day, startHour, startMinute, endHour, endMinute, breakMinutes)
        }
    }
    if !found {
        fmt.Println("No working hours configured.")
    }
    if err = rows.Err(); err != nil {
        log.Fatalf("Error after listing working hours: %v", err)
    }
}

// ListProjects lists all projects.
// It now accepts *TodoManager.
func ListProjects(tm *TodoManager) {
    rows, err := tm.db.Query("SELECT id, name FROM projects ORDER BY name ASC")
    if err != nil {
        log.Fatalf("Error listing projects: %v", err)
    }
    defer rows.Close()

    fmt.Println("----------------------------")
    fmt.Println("  ID    Project")
    fmt.Println("----------------------------")
    found := false
    for rows.Next() {
        found = true
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            log.Printf("Error scanning project: %v", err)
            continue
        }
        fmt.Printf("  %-5d %s%s%s\n", id, fg_green, name, style_reset)
    }
    if !found {
        fmt.Println("No projects found.")
    }
    if err = rows.Err(); err != nil {
        log.Fatalf("Error after listing projects: %v", err)
    }
}

// ListContexts lists all contexts.
// It now accepts *TodoManager.
func ListContexts(tm *TodoManager) {
    rows, err := tm.db.Query("SELECT id, name FROM contexts ORDER BY name ASC")
    if err != nil {
        log.Fatalf("Error listing contexts: %v", err)
    }
    defer rows.Close()

    fmt.Println("----------------------------")
    fmt.Println("  ID    Context")
    fmt.Println("----------------------------")
    found := false
    for rows.Next() {
        found = true
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            log.Printf("Error scanning context: %v", err)
            continue
        }
        fmt.Printf("  %-5d %s%s%s\n", id, fg_magenta, name, style_reset)
    }
    if !found {
        fmt.Println("No contexts found.")
    }
    if err = rows.Err(); err != nil {
        log.Fatalf("Error after listing contexts: %v", err)
    }
}

// ListTags lists all tags.
// It now accepts *TodoManager.
func ListTags(tm *TodoManager) {
    rows, err := tm.db.Query("SELECT id, name FROM tags ORDER BY name ASC")
    if err != nil {
        log.Fatalf("Error listing tags: %v", err)
    }
    defer rows.Close()

    fmt.Println("----------------------------")
    fmt.Println("  ID    Tag")
    fmt.Println("----------------------------")
    found := false
    for rows.Next() {
        found = true
        var id int
        var name string
        if err := rows.Scan(&id, &name); err != nil {
            log.Printf("Error scanning tag: %v", err)
            continue
        }
        fmt.Printf("  %-5d %s%s%s\n", id, fg_blue, name, style_reset)
    }
    if !found {
        fmt.Println("No tags found.")
    }
    if err = rows.Err(); err != nil {
        log.Fatalf("Error after listing tags: %v", err)
    }
}

