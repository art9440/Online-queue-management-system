package ics

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"strings"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	novosibirsk := time.FixedZone("Asia/Novosibirsk", 7*60*60)
	comment := "Первый визит; позвонить, если опоздаю"

	got := string(Build(&domain.AppointmentCalendarDetails{
		AppointmentID:   42,
		ServiceName:     "Consultation",
		EmployeeName:    "Anna",
		EmployeeSurname: "Petrova",
		BranchName:      "Central Branch",
		BranchAddress:   "10 Lenina street, Novosibirsk",
		StartTime:       time.Date(2030, 1, 10, 10, 0, 0, 0, novosibirsk),
		EndTime:         time.Date(2030, 1, 10, 10, 30, 0, 0, novosibirsk),
		Comment:         &comment,
	}, time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)))

	unfolded := strings.ReplaceAll(got, "\r\n ", "")

	for _, want := range []string{
		"BEGIN:VCALENDAR\r\n",
		"UID:appointment-42@online-queue\r\n",
		"DTSTAMP:20300101T120000Z\r\n",
		"DTSTART:20300110T030000Z\r\n",
		"DTEND:20300110T033000Z\r\n",
		"SUMMARY:Consultation\r\n",
		`LOCATION:Central Branch\, 10 Lenina street\, Novosibirsk` + "\r\n",
		`DESCRIPTION:Мастер: Anna Petrova\nКомментарий: Первый визит\; позвонить\, если опоздаю` + "\r\n",
		"TRIGGER:-PT1H\r\n",
		"END:VCALENDAR\r\n",
	} {
		if !strings.Contains(unfolded, want) {
			t.Errorf("calendar does not contain %q\n%s", want, got)
		}
	}

	for _, line := range strings.Split(strings.TrimSuffix(got, "\r\n"), "\r\n") {
		if len(line) > maxLineOctets {
			t.Errorf("line longer than %d octets: %q", maxLineOctets, line)
		}
	}
}

func TestFoldKeepsUTF8Intact(t *testing.T) {
	line := "DESCRIPTION:" + strings.Repeat("я", 100)

	folded := fold(line)

	for _, part := range strings.Split(folded, "\r\n") {
		if len(part) > maxLineOctets {
			t.Errorf("part longer than %d octets: %d", maxLineOctets, len(part))
		}
		if !strings.HasPrefix(part, "DESCRIPTION:") && !strings.HasPrefix(part, " я") {
			t.Errorf("continuation line must start with a space and a whole rune: %q", part)
		}
	}
	if strings.ReplaceAll(folded, "\r\n ", "") != line {
		t.Error("unfolding must restore the original line")
	}
}
