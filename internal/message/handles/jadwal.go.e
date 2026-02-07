package handles

import (
	"encoding/json"
	"fmt"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/utils/datetime"
	"kano/internal/utils/messageutil"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
)

const semsId = 2

func jam(t time.Time) string {
	return fmt.Sprintf("%02d:%02d", t.Hour()+7, t.Minute())
}

func classifyWeeks(baseWeek time.Time, dates [][2]time.Time) map[string][]int {
	if len(dates) == 0 {
		return nil
	}

	res := map[string][]int{}

	for _, date := range dates {
		dayName := daysOfWeek[date[0].Weekday()]

		weekStart := datetime.StartOfWeek(date[0])
		diffDays := (weekStart.UnixMilli() - baseWeek.UnixMilli()) / 86400000

		weekIndex := int(math.Floor(float64(diffDays/7)) + 1)
		str := fmt.Sprintf("%s %s-%s", dayName, jam(date[0]), jam(date[1]))

		res[str] = append(res[str], weekIndex)
	}

	return res
}

func removeDuplicate[T comparable](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func simplifyNumber(nums []int) string {
	if len(nums) == 0 {
		return ""
	}

	nums = removeDuplicate(nums)

	start := nums[0]
	end := nums[0]

	res := []string{}
	for i, num := range nums {
		if i == 0 {
			continue
		}
		if end+1 == num {
			end = num
		} else {
			if start == end {
				res = append(res, fmt.Sprintf("%d", start))
			} else {
				res = append(res, fmt.Sprintf("%d-%d", start, end))
			}

			start = num
			end = num
		}
	}

	if start == end {
		res = append(res, fmt.Sprintf("%d", start))
	} else {
		res = append(res, fmt.Sprintf("%d-%d", start, end))
	}

	return strings.Join(res, ", ")
}

var daysOfWeek = []string{
	"Minggu",
	"Senin",
	"Selasa",
	"Rabu",
	"Kamis",
	"Jumat",
	"Sabtu",
}

var jakarta, _ = time.LoadLocation("Asia/Jakarta")

func Jadwal(ctx *messageutil.MessageContext) error {
	db := database.GetInstance()
	args := ctx.Parser.Args
	if len(args) == 0 {
		ctx.QuoteReply("Berikan kode matkul dan/atau kelas (Contoh: ET2202, ET2202-01)")
		return nil
	}

	classCode := args[0].Content.Data
	split := strings.Split(classCode, "-")
	if len(split) > 2 {
		ctx.QuoteReply("Format kode kelas salah (Contoh yang benar: ET2202, ET2202-01)")
		return nil
	}

	var repMsg strings.Builder

	var filterHari []int
	hari, ok := ctx.Parser.NamedArgs["hari"]
	if ok {
		if len(hari) > 0 {
			data := hari[0].Content.Data
			split := strings.SplitSeq(data, ",")
			for spl := range split {
				idx := -1
				for i, h := range daysOfWeek {
					if strings.ToLower(h) == strings.TrimSpace(strings.ToLower(spl)) {
						idx = i
						break
					}
				}
				if idx != -1 && !slices.Contains(filterHari, idx) {
					filterHari = append(filterHari, idx)
				} else {
					if idx == -1 {
						fmt.Fprintf(&repMsg, "_Invalid weekday %s_\n", spl)
					} else {
						fmt.Fprintf(&repMsg, "_Duplicated weekday %s_\n", spl)
					}
				}
			}
		}
	}

	var filterJam [][2]uint64 // Start and end
	jams, ok := ctx.Parser.NamedArgs["jam"]
	if ok {
		if len(jams) > 0 {
			data := jams[0].Content.Data
			split := strings.SplitSeq(data, ",")
			for spl := range split {
				s := strings.Split(spl, "-")
				if len(s) > 2 {
					fmt.Fprintf(&repMsg, "_%s has invalid separator count_\n", spl)
					continue
				}
				start, err := strconv.ParseUint(s[0], 10, 0)
				if err != nil {
					fmt.Fprintf(&repMsg, "_%s is not a number_\n", s[0])
					continue
				}
				if len(s) == 2 {
					end, err := strconv.ParseUint(s[1], 10, 0)
					if err != nil {
						fmt.Fprintf(&repMsg, "_%s is not a number_\n", s[0])
						end = start
					}
					filterJam = append(filterJam, [2]uint64{start, end})
				} else {
					filterJam = append(filterJam, [2]uint64{start, start})
				}
			}
		}
	}

	var filterProdi []uint64
	prodis, ok := ctx.Parser.NamedArgs["prodi"]
	if ok {
		if len(prodis) > 0 {
			data := prodis[0].Content.Data
			split := strings.SplitSeq(data, ",")
			for spl := range split {
				prodi, err := strconv.ParseUint(spl, 10, 0)
				if err != nil {
					fmt.Fprintf(&repMsg, "_%s is not a number_\n", spl)
					continue
				}
				if !slices.Contains(filterProdi, prodi) {
					filterProdi = append(filterProdi, prodi)
				}
			}
		}
	}

	var filterFakul []string
	facs, ok := ctx.Parser.NamedArgs["fakultas"]
	if ok {
		if len(facs) > 0 {
			data := facs[0].Content.Data
			split := strings.SplitSeq(data, ",")
			for spl := range split {
				if !slices.Contains(filterFakul, spl) {
					filterFakul = append(filterFakul, spl)
				}
			}
		}
	}
	_, strictFilter := ctx.Parser.NamedArgs["strict"]

	subjCode := split[0]
	if len(subjCode) > 6 {
		ctx.QuoteReply("Panjang kode matkul lebih dari 6 (Contoh yang benar: ET2202-01)")
		return nil
	}
	subjCode = strings.ToUpper(subjCode)

	var classes []models.SubjectClass
	q := db.
		Table("subject_class AS sc").
		Joins(`JOIN subject_schedule ss ON ss.id = sc."subject_schedule_id"`).
		Joins(`JOIN subject s ON s.id = ss."subject_id"`).
		Where(`s.code LIKE ?`, subjCode+"%").
		Where(`ss."semester_id" = ?`, semsId).
		Preload("AvailableAtMajor").
		Preload("SubjectSchedule.Subject").
		Preload("Lecturers.Lecturer").
		Preload("Schedules").
		Preload("Constraint.Majors.ConstraintMajor.Major")

	if len(split) == 2 {
		classNum, err := strconv.ParseUint(split[1], 10, 0)
		if err != nil {
			ctx.QuoteReply("Nomor kelas bukan angka (Contoh yang benar: ET2202-01)")
			return nil
		}
		if classNum < 1 {
			ctx.QuoteReply("Nomor kelas harus lebih besar dari 0 (Contoh yang benar: ET2202-01)")
			return nil
		}

		q = q.Where(`sc.number = ?`, classNum)
	}

	tx := q.Find(&classes)
	if tx.Error != nil {
		ctx.QuoteReply("[1] Internal error: %s", tx.Error)
		return nil
	}

	if ctx.IsSenderSame(config.GetConfig().OwnerJID) && ctx.GetChat().Server != types.GroupServer {
		mar, _ := json.MarshalIndent(classes, "", " ")
		ctx.QuoteReply("%s", mar)
	}

	for _, theClass := range classes {
		var schedules []models.ClassSchedule
		tx = db.Where(`"subject_class_id" = ?`, theClass.ID).Find(&schedules)
		if tx.Error != nil {
			ctx.QuoteReply("[2] Internal error: %s", tx.Error)
			return nil
		}

		var schedStr strings.Builder
		if len(schedules) > 0 {
			// Filter input
			filterSatisfied := false // Default true, bakal false tergantung filter
			for _, sched := range schedules {
				var hariSatisfied, jamSatisfied, constraintSatisfied bool
				// Filter hari
				weekday := sched.Start.Weekday()
				if len(filterHari) > 0 {
					if slices.Contains(filterHari, int(weekday)) { // Jadwal masuk dalam filter
						hariSatisfied = true
					}
				} else { // Ga ada filter hari
					hariSatisfied = true
				}

				// Filter jam
				if len(filterJam) > 0 {
					// Apakah jadwal ini memiliki jam yang memenuhi spesifikasi user
					exist := slices.ContainsFunc(filterJam, func(j [2]uint64) bool {
						start := sched.Start.In(jakarta)
						end := sched.End.In(jakarta)

						rangeStart := time.Date(
							start.Year(), start.Month(), start.Day(),
							int(j[0]), 0, 0, 0, jakarta,
						)

						rangeEnd := time.Date(
							start.Year(), start.Month(), start.Day(),
							int(j[1]), 0, 0, 0, jakarta,
						)

						// overlap check
						return start.Before(rangeEnd) && end.After(rangeStart)
					})

					if exist {
						jamSatisfied = true
					}
				} else { // Ga ada filter jam
					jamSatisfied = true
				}

				// Filter batasan prodi atau fakul
				if len(filterProdi) > 0 || len(filterFakul) > 0 {
					constraint := theClass.Constraint
					if constraint == nil { // Kelas ga ada batasan
						constraintSatisfied = !strictFilter // Jika strict, maka filterSatisfied = false, artinya harus ada constraint
						continue
					} else {
						// Kalau ada batasan prodi
						if len(constraint.Majors) > 0 {
							exist := false
							for _, maj := range constraint.Majors {
								cm := maj.ConstraintMajor
								if cm == nil {
									continue
								}
								m := cm.Major
								if m == nil {
									continue
								}
								id := m.ID
								if slices.Contains(filterProdi, uint64(id)) {
									exist = true
								}
							}
							if exist {
								constraintSatisfied = true
								break
							}
						} else if len(constraint.Faculties) > 0 {
							exist := slices.ContainsFunc(constraint.Faculties, func(e string) bool {
								return slices.ContainsFunc(filterFakul, func(d string) bool {
									return strings.EqualFold(d, e)
								})
							})
							if exist {
								constraintSatisfied = true
							}
						} else { // Kalau ga ada? Ini kasus aneh, aku anggap tidak satisfied
							constraintSatisfied = false
						}
					}
				} else { // Ga ada filter prodi/fakultas
					constraintSatisfied = true
				}

				filterSatisfied = filterSatisfied || (hariSatisfied && jamSatisfied && constraintSatisfied)
			}
			if !filterSatisfied {
				continue // Skip kelas
			}

			baseWeekStart := datetime.StartOfWeek(schedules[0].Start)

			var quiz, uts, uas string
			var scheddates [][2]time.Time

			for _, sched := range schedules {
				weekStart := datetime.StartOfWeek(sched.Start)
				diffDays := math.Floor(float64(weekStart.UnixMilli() - baseWeekStart.UnixMilli()/86400000/7))
				str := fmt.Sprintf("[Pekan %d] %s %s-%s", int(diffDays)+1, sched.Start.Weekday().String(), jam(sched.Start), jam(sched.End))

				switch sched.Activity {
				case models.ScheduleActivityQuiz:
					quiz += str
				case models.ScheduleActivityMidterm:
					uts += str
				case models.ScheduleActivityFinal:
					uas += str
				default:
					scheddates = append(scheddates, [2]time.Time{sched.Start, sched.End})
				}
			}

			classified := classifyWeeks(baseWeekStart, scheddates)
			uhh := []struct {
				Name  string
				Weeks []int
			}{}
			for name, weeks := range classified {
				uhh = append(uhh, struct {
					Name  string
					Weeks []int
				}{name, weeks})
			}
			slices.SortFunc(uhh, func(a, b struct {
				Name  string
				Weeks []int
			}) int {
				return len(b.Weeks) - len(a.Weeks)
			})

			for i, u := range uhh {
				fmt.Fprintf(&schedStr, "- %s: Pekan %s", u.Name, simplifyNumber(u.Weeks))
				if i != len(uhh)-1 {
					schedStr.WriteString("\n")
				}
			}
		}

		lecturers := make([]string, len(theClass.Lecturers))
		for i, lecturer := range theClass.Lecturers {
			lecturers[i] = fmt.Sprintf("- %s", lecturer.Lecturer.Name)
		}

		var jadwal string = "- Ga ada."
		if len(schedStr.String()) > 0 {
			jadwal = schedStr.String()
		}

		var dosen string = "- Somehow ga ada (special case)"
		if len(lecturers) > 0 {
			dosen = strings.Join(lecturers, "\n")
		}

		var constraintStr strings.Builder
		if c := theClass.Constraint; c != nil {
			constraintStr.WriteString("Batasan: ")
			if len(c.Stratas) > 0 {
				fmt.Fprintf(&constraintStr, "\n- Strata %s", strings.Join(c.Stratas, ", "))
			}
			if len(c.Campuses) > 0 {
				fmt.Fprintf(&constraintStr, "\n- Kampus %s", strings.Join(c.Campuses, ", "))
			}
			if len(c.Faculties) > 0 {
				fmt.Fprintf(&constraintStr, "\n- Fakultas %s", strings.Join(c.Faculties, ", "))
			}
			if len(c.Majors) > 0 {
				pprodi := make([]string, len(c.Majors))
				for i, m := range c.Majors {
					cm := m.ConstraintMajor
					if cm == nil {
						continue
					}
					mj := cm.Major
					if mj == nil {
						pprodi[i] = fmt.Sprintf("%s", cm.AddonData)
						continue
					}
					pprodi[i] = fmt.Sprintf("%s %s", mj.Name, cm.AddonData)
				}
				fmt.Fprintf(&constraintStr, "\n- Prodi %s", strings.Join(pprodi, ", "))
			}
			if constraintStr.Len() < 10 {
				constraintStr.WriteString("Tidak ada")
			}
		}

		fmt.Fprintf(
			&repMsg,
			"*%s-%02d - %s*\nID Kelas: %d\n%s\nDosen:\n%s\nRingkasan jadwal:\n%s",
			theClass.SubjectSchedule.Subject.Code, theClass.Number, theClass.SubjectSchedule.Subject.Name,
			theClass.ID, constraintStr.String(), dosen,
			jadwal,
		)

		repMsg.WriteString("\n\n")
	}

	if repMsg.String() == "" {
		ctx.QuoteReply("Ga ketemu")
	} else {
		ctx.QuoteReply("%s", strings.TrimSpace(repMsg.String()))
	}

	return nil
}
