package six

import "kano/internal/utils/six/schedules"

func GetAllSchedules() ([]schedules.SemesterSubject, error) {
	return schedules.GetSchedules()
}
