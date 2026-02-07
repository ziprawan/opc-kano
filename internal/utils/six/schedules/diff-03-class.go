package schedules

import (
	"kano/internal/database/models"
	"slices"
)

// Self note
// classDiff : The diff object
// classMod  : New class data
// dbClass   : Old class data

func linksDiff(classDiff *ClassDiff, classMod Class, dbClass models.SubjectClass) {
	if classDiff == nil {
		return
	}

	if classMod.Links.EdunexClassId != dbClass.EdunexClassID || classMod.Links.Teams != dbClass.TeamsLink.String {
		classDiff.Links.HasDiff = true
		classDiff.Links.Before.EdunexClassId = dbClass.EdunexClassID
		classDiff.Links.After.EdunexClassId = classMod.Links.EdunexClassId
		classDiff.Links.Before.Teams = dbClass.TeamsLink.String
		classDiff.Links.After.Teams = classMod.Links.Teams
	}
}

func constraintDiff(classDiff *ClassDiff, classMod Class, dbClass models.SubjectClass) {
	if classDiff == nil {
		return
	}

	new := classMod.Constraints
	old := modelConstraintToScheduleConstraint(dbClass.Constraint)

	valid := new.Valid == old.Valid
	fac := compareArr(new.Constraint.Faculties, old.Constraint.Faculties)
	str := compareArr(new.Constraint.Stratas, old.Constraint.Stratas)
	cam := compareArr(new.Constraint.Campuses, old.Constraint.Campuses)
	sem := compareArr(new.Constraint.Semesters, old.Constraint.Semesters)
	maj := compareArr(new.Constraint.Majors, old.Constraint.Majors)
	oth := compareArr(new.Constraint.Others, old.Constraint.Others)

	diff := !valid || !fac || !str || !cam || !sem || !maj || !oth
	if diff {
		classDiff.Constraints.HasDiff = true
		classDiff.Constraints.Before = old
		classDiff.Constraints.After = new
	}
}

func lecturerDiff(classDiff *ClassDiff, classMod Class, dbClass models.SubjectClass) {
	if classDiff == nil {
		return
	}

	new := classMod.Lecturers
	old := modelLecturerToString(dbClass.Lecturers)

	added := make([]string, 0, len(new))
	for _, newLec := range new {
		if !slices.Contains(old, newLec) {
			added = append(added, newLec)
		}
	}
	removed := make([]string, 0, len(old))
	for _, oldLec := range old {
		if !slices.Contains(new, oldLec) {
			removed = append(removed, oldLec)
		}
	}

	classDiff.AddedLecturers = added
	classDiff.RemovedLecturers = removed
}

func scheduleDiff(classDiff *ClassDiff, classMod Class, dbClass models.SubjectClass) error {
	// Find removed
	removed := make([]models.ClassSchedule, 0, len(dbClass.Schedules))
	for _, dbSched := range dbClass.Schedules {
		if !slices.ContainsFunc(classMod.Schedules, func(a Schedule) bool {
			return a.Start.Equal(dbSched.Start) &&
				a.End.Equal(dbSched.End) &&
				string(a.Activity) == string(dbSched.Activity) &&
				string(a.Method) == string(dbSched.Method)
		}) {
			removed = append(removed, dbSched)
		}
	}

	// Find added and exists (for modified one)
	added := make([]Schedule, 0, len(classMod.Schedules))
	exists := make([]Schedule, 0, len(classMod.Schedules))
	for _, modSched := range classMod.Schedules {
		if !slices.ContainsFunc(dbClass.Schedules, func(a models.ClassSchedule) bool {
			return a.Start.Equal(modSched.Start) &&
				a.End.Equal(modSched.End) &&
				string(a.Activity) == string(modSched.Activity) &&
				string(a.Method) == string(modSched.Method)
		}) {
			added = append(added, modSched)
		} else {
			exists = append(exists, modSched)
		}
	}

	// classDiff.ModifiedSchedules = make([]ScheduleDiff, 0, len(exists))
	// for i, schedMod := range exists {
	// 	var schedDiff ScheduleDiff

	// 	idx := slices.IndexFunc(dbClass.Schedules, func(a models.ClassSchedule) bool { return a.Start.Equal(schedMod.Start) && a.End.Equal(schedMod.End) })
	// 	if idx == -1 {
	// 		return fmt.Errorf("schedDiff idx %d: invalid index -1", i)
	// 	}
	// 	dbSched := dbClass.Schedules[idx]

	// 	schedDiff.ID = dbSched.ID
	// 	if string(schedMod.Activity) != string(dbSched.Activity) {
	// 		schedDiff.Activity = varDiff[Activity]{
	// 			Before:  Activity(dbSched.Activity),
	// 			After:   schedMod.Activity,
	// 			HasDiff: true,
	// 		}
	// 	}
	// 	if string(schedMod.Method) != string(dbSched.Method) {
	// 		schedDiff.Method = varDiff[Method]{
	// 			Before:  Method(dbSched.Method),
	// 			After:   schedMod.Method,
	// 			HasDiff: true,
	// 		}
	// 	}
	// 	handleRoom(&schedDiff, schedMod, dbSched)

	// 	// Only append if there is any difference!
	// 	anyDiff := schedDiff.Activity.HasDiff ||
	// 		schedDiff.Method.HasDiff ||
	// 		len(schedDiff.AddedRooms) > 0 ||
	// 		len(schedDiff.RemovedRooms) > 0

	// 	if anyDiff {
	// 		classDiff.ModifiedSchedules = append(classDiff.ModifiedSchedules, schedDiff)
	// 	}
	// }

	classDiff.AddedSchedules = added
	classDiff.RemovedSchedules = removed

	return nil
}

func handleRoom(schedDiff *ScheduleDiff, schedMod Schedule, dbSched models.ClassSchedule) {
	if schedDiff == nil {
		return
	}

	new := schedMod.Rooms
	old := modelRoomToString(dbSched.Rooms)

	added := make([]string, 0, len(new))
	for _, newR := range new {
		if !slices.Contains(old, newR) {
			added = append(added, newR)
		}
	}
	removed := make([]string, 0, len(old))
	for _, oldR := range old {
		if !slices.Contains(new, oldR) {
			removed = append(removed, oldR)
		}
	}

	schedDiff.AddedRooms = added
	schedDiff.RemovedRooms = removed
}

// Other func and type that only help
type hasName interface {
	struct{ Name string }
}

func modelLecturerToString(lecturers []models.Lecturer) []string {
	res := make([]string, len(lecturers))
	for i := range lecturers {
		res[i] = lecturers[i].Name
	}

	return res
}

func modelRoomToString(rooms []models.Room) []string {
	res := make([]string, len(rooms))
	for i := range rooms {
		res[i] = rooms[i].Name
	}

	return res
}

func modelConstraintToScheduleConstraint(model *models.ClassConstraint) (res NullConstraint) {
	if model == nil {
		// There is nothing I can do
		return
	}

	stratas := make([]Strata, len(model.Stratas))
	for i := range model.Stratas {
		stratas[i] = Strata(model.Stratas[i])
	}

	campuses := make([]Campus, len(model.Campuses))
	for i := range model.Campuses {
		campuses[i] = Campus(model.Campuses[i])
	}

	semesters := make([]uint, len(model.Semester))
	for i := range model.Semester {
		semesters[i] = uint(model.Semester[i])
	}

	majors := make([]MajorConstraint, len(model.Majors))
	for i := range model.Majors {
		majors[i] = MajorConstraint{
			ID:    uint(model.Majors[i].MajorID.Int32),
			Addon: model.Majors[i].AddonData,
		}
	}

	res.Valid = true
	res.Constraint = Constraint{
		Faculties: model.Faculties,
		Others:    model.Other,
		Stratas:   stratas,
		Campuses:  campuses,
		Semesters: semesters,
		Majors:    majors,
	}

	return
}

func compareArr[T comparable](arr1 []T, arr2 []T) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	maps := map[T]int{}
	for _, a := range arr1 {
		maps[a]++
	}
	for _, b := range arr2 {
		maps[b]--
	}

	for _, l := range maps {
		if l != 0 {
			return false
		}
	}

	return true
}
