package schedules

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

func getEdunexClassIdByPath(url *url.URL) (sql.NullInt32, error) {
	empty := sql.NullInt32{}
	pathSplit := strings.Split(url.Path, "/")
	if len(pathSplit) <= 3 {
		return empty, fmt.Errorf("invalid pathsplit length, expected more than 3, got %d", len(pathSplit))
	}
	slices.Reverse(pathSplit)

	classCtx := pathSplit[0]
	semsCtx := pathSplit[2]

	subjectCode, classNumStr, ok := strings.Cut(classCtx, "-")
	if !ok {
		return empty, fmt.Errorf("invalid classCtx %s: cannot be splitted by \"-\"", classCtx)
	}
	classNum, err := strconv.ParseUint(classNumStr, 10, 0)
	if err != nil {
		return empty, fmt.Errorf("unable to convert classNumStr as uint: %s", classNumStr)
	}

	ed := models.EdunexClassId{
		SemsCtx:     semsCtx,
		ClassNum:    uint(classNum),
		SubjectCode: subjectCode,
	}
	db := database.GetInstance()
	tx := db.Where(&ed).First(&ed)
	if err = tx.Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return empty, err
		}

		toInsert := models.EdunexClassId{
			SemsCtx:     semsCtx,
			SubjectCode: subjectCode,
			ClassNum:    uint(classNum),
		}

		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Head(url.String())
		if err != nil {
			if !errors.Is(err, context.DeadlineExceeded) {
				return empty, err
			}
		}
		if resp != nil {
			defer resp.Body.Close()
		}

		if resp != nil && resp.Request.URL != nil {
			redirUrl := resp.Request.URL
			redirSplit := strings.Split(redirUrl.Path, "/")
			slices.Reverse(redirSplit)
			if len(redirSplit) <= 2 {
				return empty, fmt.Errorf("invalid redirect path: %s", redirUrl.Path)
			}
			theClassId, err := strconv.ParseUint(redirSplit[1], 10, 0)
			if err != nil {
				return empty, fmt.Errorf("cannot convert class id into uint: %s", redirSplit[1])
			}
			toInsert.ClassId = sql.NullInt32{
				Int32: int32(theClassId),
				Valid: true,
			}
		}

		tx = db.Create(&toInsert)
		return toInsert.ClassId, tx.Error
	} else {
		return ed.ClassId, nil
	}
}
