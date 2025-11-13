package utils

import (
	"fmt"
	"strings"
	"time"
)

const (
	scheduleStartMinutes = 10 * 60 // 10:00
	scheduleEndMinutes   = 26 * 60 // 02:00 (next day)
	minutesPerDay        = 24 * 60
	defaultSlotMinutes   = 60
	extendedSlotMinutes  = 90
)

// slotDurationForType returns the slot duration in minutes for a cancha type.
func slotDurationForType(canchaType string) int {
	t := strings.ToLower(strings.TrimSpace(canchaType))
	if t == "padel" || t == "tenis" || t == "paddle" {
		return extendedSlotMinutes
	}
	return defaultSlotMinutes
}

// NormalizeSlotMinutes converts an HH:MM string into minutes, handling the 10:00-02:00 window.
func NormalizeSlotMinutes(timeStr string) (int, error) {
	layout := "15:04"
	parsed, err := time.Parse(layout, timeStr)
	if err != nil {
		return 0, fmt.Errorf("invalid time format: %w", err)
	}

	minutes := parsed.Hour()*60 + parsed.Minute()
	if minutes < scheduleStartMinutes {
		minutes += minutesPerDay
	}
	return minutes, nil
}

// minutesToClock converts absolute minutes back to an HH:MM string.
func minutesToClock(totalMinutes int) string {
	minutes := totalMinutes % minutesPerDay
	if minutes < 0 {
		minutes += minutesPerDay
	}

	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%02d:%02d", hours, mins)
}

// IntervalsOverlap reports whether two ranges [start, end) overlap.
func IntervalsOverlap(startA, endA, startB, endB int) bool {
	return startA < endB && startB < endA
}

// EnsureValidSlot validates and normalizes a slot according to cancha type rules.
// Returns normalized start, calculated end, and duration in minutes.
func EnsureValidSlot(canchaType, startTime, providedEndTime string) (string, string, int, error) {
	duration := slotDurationForType(canchaType)

	startMinutes, err := NormalizeSlotMinutes(startTime)
	if err != nil {
		return "", "", 0, err
	}

	if startMinutes < scheduleStartMinutes || startMinutes >= scheduleEndMinutes {
		return "", "", 0, fmt.Errorf("start time must be between 10:00 and 01:59")
	}

	if (startMinutes-scheduleStartMinutes)%duration != 0 {
		return "", "", 0, fmt.Errorf("start time must align with %d-minute slots for this cancha", duration)
	}

	expectedEnd := startMinutes + duration
	if expectedEnd > scheduleEndMinutes {
		return "", "", 0, fmt.Errorf("selected slot exceeds closing time (02:00)")
	}

	if providedEndTime != "" {
		endMinutes, err := NormalizeSlotMinutes(providedEndTime)
		if err != nil {
			return "", "", 0, err
		}
		if endMinutes != expectedEnd {
			return "", "", 0, fmt.Errorf("end time must be %s for this cancha type", minutesToClock(expectedEnd))
		}
	}

	return minutesToClock(startMinutes), minutesToClock(expectedEnd), duration, nil
}
