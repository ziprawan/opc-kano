package schedules

import (
	"fmt"
	"kano/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func handleModifiedSchedules(dbx *gorm.DB, modScheds []ScheduleDiff) error {
	addedRoomTotal := 0
	removedRoomTotal := 0
	for _, mSched := range modScheds {
		// Check all the modified schedule fields
		hasChange := false
		theChange := models.ClassSchedule{}

		if mSched.Activity.HasDiff {
			hasChange = true
			theChange.Activity = models.ScheduleActivity(mSched.Activity.After)
		}
		if mSched.Method.HasDiff {
			hasChange = true
			theChange.Method = models.ScheduleMethod(mSched.Method.After)
		}
		if hasChange {
			tx := dbx.Where("id = ?", mSched.ID).Updates(&theChange)
			if tx.Error != nil {
				return fmt.Errorf("failed to update sched id %d: %s", mSched.ID, tx.Error)
			}
		}

		addedRoomTotal += len(mSched.AddedRooms)
		removedRoomTotal += len(mSched.RemovedRooms)
	}
	roomsToInsert := make([]models.RoomInClass, 0, addedRoomTotal)
	roomsToRemove := make([][2]uint, 0, removedRoomTotal)
	for _, mSched := range modScheds {
		for _, ar := range mSched.AddedRooms {
			id := findRoomId(ar)
			if id == 0 {
				return fmt.Errorf("unable to find room id for name=%q", ar)
			}
			roomsToInsert = append(roomsToInsert, models.RoomInClass{
				RoomID:          id,
				ClassScheduleID: mSched.ID,
			})
		}
		for _, rr := range mSched.RemovedRooms {
			id := findRoomId(rr)
			if id == 0 {
				return fmt.Errorf("unable to find room id for name=%q", rr)
			}
			roomsToRemove = append(roomsToRemove, [2]uint{mSched.ID, id})
		}
	}

	// Insert room_in_class
	if len(roomsToInsert) > 0 {
		tx := dbx.Clauses(clause.OnConflict{DoNothing: true}).Create(&roomsToInsert)
		if tx.Error != nil {
			return fmt.Errorf("failed to insert room_in_class: %s", tx.Error)
		}
	}

	// Delete room_in_class
	if len(roomsToRemove) > 0 {
		tx := dbx.Where("(class_schedule_id, room_id) IN ?", roomsToRemove).Delete(&models.RoomInClass{})
		if tx.Error != nil {
			return fmt.Errorf("failed to remove room_in_class: %s", tx.Error)
		}
	}

	return nil
}
