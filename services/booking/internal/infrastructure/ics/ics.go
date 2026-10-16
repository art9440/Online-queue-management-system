// Package ics renders appointments as iCalendar (RFC 5545) files that any calendar app can import.
package ics

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	utcLayout     = "20060102T150405Z"
	maxLineOctets = 75
)

// Build renders a single-event calendar with a reminder one hour before the appointment.
func Build(details *domain.AppointmentCalendarDetails, now time.Time) []byte {
	summary := details.ServiceName

	description := fmt.Sprintf("Мастер: %s %s", details.EmployeeName, details.EmployeeSurname)
	if details.Comment != nil && *details.Comment != "" {
		description += "\nКомментарий: " + *details.Comment
	}

	location := details.BranchName
	if details.BranchAddress != "" {
		location += ", " + details.BranchAddress
	}

	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//Online Queue//Booking//RU",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		fmt.Sprintf("UID:appointment-%d@online-queue", details.AppointmentID),
		"DTSTAMP:" + now.UTC().Format(utcLayout),
		"DTSTART:" + details.StartTime.UTC().Format(utcLayout),
		"DTEND:" + details.EndTime.UTC().Format(utcLayout),
		"SUMMARY:" + escapeText(summary),
		"LOCATION:" + escapeText(location),
		"DESCRIPTION:" + escapeText(description),
		"BEGIN:VALARM",
		"ACTION:DISPLAY",
		"TRIGGER:-PT1H",
		"DESCRIPTION:" + escapeText(summary),
		"END:VALARM",
		"END:VEVENT",
		"END:VCALENDAR",
	}

	var b strings.Builder
	for _, line := range lines {
		b.WriteString(fold(line))
		b.WriteString("\r\n")
	}

	return []byte(b.String())
}

// escapeText escapes a TEXT value per RFC 5545 section 3.3.11.
func escapeText(s string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		";", `\;`,
		",", `\,`,
		"\r\n", `\n`,
		"\n", `\n`,
		"\r", `\n`,
	).Replace(s)
}

// fold splits a content line into lines of at most 75 octets without breaking UTF-8 characters
// (RFC 5545 section 3.1). Continuation lines start with a single space, which counts towards the limit.
func fold(line string) string {
	if len(line) <= maxLineOctets {
		return line
	}

	var b strings.Builder
	limit := maxLineOctets
	width := 0

	for _, r := range line {
		size := utf8.RuneLen(r)
		if width+size > limit {
			b.WriteString("\r\n ")
			limit = maxLineOctets - 1
			width = 0
		}
		b.WriteRune(r)
		width += size
	}

	return b.String()
}
