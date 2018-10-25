package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

//
func writeJSON(w http.ResponseWriter, data interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}

//
func writeError(w http.ResponseWriter, message string, code int) error {
	return writeJSON(w, map[string][]map[string]string{
		"errors": []map[string]string{
			map[string]string{"message": message},
		},
	}, code)
}

type ClockTime struct {
	hour	int
	minute	int
}

type Schedule []ClockTime

type TimeSlice []time.Time

func timeFromClockTime(ct ClockTime) time.Time {
	tz, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(tz)
	year := now.Year()
	month := now.Month()
	day := now.Day()

	return time.Date(year, month, day, ct.hour, ct.minute, 0, 0, tz)
}

func timeSliceFromSchedule(s Schedule) TimeSlice {
	ts := make(TimeSlice, len(s))
	for i, ct := range s {
		ts[i] = timeFromClockTime(ct)
	}

	return ts
}

func timesAfter(ts TimeSlice, comp time.Time, limit int) TimeSlice {
	tsf := make(TimeSlice, 0)
	for _, t := range ts {
		if t.After(comp) {
			tsf = append(tsf, t)
			if len(tsf) >= limit { break }
		}
	}

	return tsf
}

func formatTimeSlice(ts TimeSlice, layout string) []string {
	tsm := make([]string, len(ts))
	for i, t := range ts {
		tsm[i] = t.Format(layout)
	}

	return tsm
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/slack/", func(w http.ResponseWriter, req *http.Request) {
		if err := req.ParseForm(); err != nil {
			writeError(w, "Error parsing form", 500)
			return
		}

		location := strings.ToLower(req.Form.Get("text"))

		var schedule Schedule

		switch location {
		case "bva":
			schedule = Schedule{
				{ hour: 7,  minute: 20 },
				{ hour: 8,  minute: 25 },
				{ hour: 9,  minute: 30 },
				{ hour: 10, minute: 35 },
				{ hour: 11, minute: 40 },
				{ hour: 12, minute: 45 },
				{ hour: 13, minute: 50 },
				{ hour: 14, minute: 55 },
				{ hour: 16, minute: 0  },
				{ hour: 17, minute: 5  },
			}
		case "vaco":
			schedule = Schedule{
				{ hour: 7,  minute: 0  },
				{ hour: 8,  minute: 5  },
				{ hour: 9,  minute: 10 },
				{ hour: 10, minute: 15 },
				{ hour: 11, minute: 20 },
				{ hour: 12, minute: 25 },
				{ hour: 13, minute: 30 },
				{ hour: 14, minute: 35 },
				{ hour: 15, minute: 40 },
				{ hour: 16, minute: 45 },
				{ hour: 17, minute: 50 },
			}
		default:
			writeError(w, "Shuttle location not found", 500)
			return
		}

		tz, _ := time.LoadLocation("America/New_York")
		now := time.Now().In(tz)
		times := timesAfter(timeSliceFromSchedule(schedule), now, 3)
		fmtdTimes := formatTimeSlice(times, "3:04PM")

		result := fmt.Sprintf("Upcoming departures: %s", strings.Join(fmtdTimes, ", "))

		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(200)
		w.Write([]byte(result))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		writeError(w, "No such page", 404)
	})

	panic(http.ListenAndServe(":2839", mux))
}
