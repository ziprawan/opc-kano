package schedules

import (
	"fmt"
	"kano/internal/database/models"
	"slices"
)

// See vars.go for related variables

func initRooms(scheds []SemesterSubject) error {
	strRooms := []string{}
	for _, sched := range scheds {
		for _, subject := range sched.Subjects {
			for _, class := range subject.Classes {
				for _, schedule := range class.Schedules {
					for _, room := range schedule.Rooms {
						if !slices.Contains(strRooms, room) {
							strRooms = append(strRooms, room)
						}
					}
				}
			}
		}
	}

	found := make([]models.Room, 0, len(strRooms))
	tx := db.Where("name IN ?", strRooms).Find(&found)
	if tx.Error != nil {
		return tx.Error
	}

	if len(found) == len(strRooms) {
		rooms = found
		return nil
	}

	missing := len(strRooms) - len(found)
	fmt.Printf("Missing %d room(s) from database\n", missing)
	toInsert := make([]models.Room, missing)
	i := 0
	for _, name := range strRooms {
		if !slices.ContainsFunc(found, func(a models.Room) bool { return a.Name == name }) {
			toInsert[i] = models.Room{Name: name}
			i++
		}
	}

	tx = db.Create(&toInsert)
	if tx.Error != nil {
		return tx.Error
	}
	fmt.Printf("Inserted %d room(s) into database\n", tx.RowsAffected)

	rooms = make([]models.Room, len(strRooms))
	i = 0
	for _, f := range found {
		rooms[i] = f
		i++
	}
	for _, in := range toInsert {
		rooms[i] = in
		i++
	}

	return nil
}

func findRoomId(roomName string) uint {
	for _, room := range rooms {
		if room.Name == roomName {
			return room.ID
		}
	}

	return 0
}
