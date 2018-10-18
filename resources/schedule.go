package resources

import "time"

//Schedule represents all schedules and availabilities for an event
type Schedule struct {
	ScheduleID  string         `json:"scheduleId"`
	CreatedDate time.Time      `json:"createdDate"`
	FixedDate   bool           `json:"fixedDate"`
	FixedPeriod bool           `json:"fixedPeriod"`
	Closed      bool           `json:"closed"`
	StartDate   time.Time      `json:"startDate"`
	EndDate     time.Time      `json:"endDate"`
	WeekDays    map[string]int `json:"weekDays"`
}
