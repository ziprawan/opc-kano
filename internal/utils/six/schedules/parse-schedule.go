package schedules

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseSchedules(col *goquery.Selection) (LastColumn, error) {
	res := LastColumn{}

	// Getting the links
	firstChild := col.Children().First()
	anchors := firstChild.Find("a")
	if anchors.Length() != 2 {
		return res, fmt.Errorf("unable to resolve class links: expected 2 anchors, got %d", anchors.Length())
	}

	// Resolve edunex id
	edunexHref, ok := anchors.Eq(0).Attr("href")
	if !ok {
		return res, fmt.Errorf("unable to get edunex href")
	}
	u, err := url.Parse(edunexHref)
	if err != nil {
		return res, err
	}
	edunexClassId, err := getEdunexClassIdByPath(u)
	if err != nil {
		return res, fmt.Errorf("getEdunexClassIdByPath: %s", err)
	}
	res.EdunexClassId = edunexClassId

	// Resolve teams link
	teamsHref, ok := anchors.Eq(1).Attr("href")
	if !ok {
		return res, fmt.Errorf("unable to get teams href")
	}
	if !strings.HasPrefix(teamsHref, "https://") {
		res.Teams = ""
		splits := strings.Split(teamsHref, "/")
		slices.Reverse(splits)
		classId, err := strconv.ParseUint(splits[1], 10, 0)
		if err != nil {
			return res, fmt.Errorf("ParseUint: Unable to parse class id from teams href: %s", teamsHref)
		}
		res.ClassId = uint(classId)
	} else {
		res.Teams = teamsHref
	}

	danger := col.Find(`span[class~="label-danger"]`)
	if danger.Length() > 0 {
		// There is a "Belum ada" text
		return res, nil
	}

	ul := col.Find("ul")
	if ul.Length() != 1 {
		return res, fmt.Errorf("expected 1 <ul>, got %d", ul.Length())
	}

	// Getting the six class id
	div := ul.Find(`div[class~="collapse"]`)
	divId, ok := div.Attr("id")
	if res.ClassId == 0 && ok {
		_, classIdStr, ok := strings.Cut(divId, "_")
		if !ok {
			return res, fmt.Errorf("invalid div id: %s", divId)
		}
		classId, err := strconv.ParseUint(classIdStr, 10, 0)
		if err != nil {
			return res, fmt.Errorf("ParseUint: unable to convert class id into uint: %s", classIdStr)
		}
		res.ClassId = uint(classId)
	}
	if res.ClassId == 0 {
		return res, fmt.Errorf("unable to obtain class id, either from teams href or div collapse")
	}

	// Getting all the schedules list hahahahahahahahhaha
	lis := ul.Find("li")
	total := 0
	lis.Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s.Contents().First()) == "#text" {
			total++
		}
	})

	res.Schedules = make([]Schedule, total)
	i := 0
	for liIdx, li := range lis.EachIter() {
		if goquery.NodeName(li.Contents().First()) != "#text" {
			continue
		}

		sched, err := parseIndividualSchedule(li)
		if err != nil {
			return res, fmt.Errorf("schedule idx %d: %s", liIdx, err)
		}
		res.Schedules[i] = sched

		i++
	}

	return res, nil
}

func parseIndividualSchedule(li *goquery.Selection) (Schedule, error) {
	res := Schedule{}

	// Raw data
	data := strings.TrimSpace(li.Text())
	dataSplit := splitTrimN(data, "/\n", 6)
	if len(dataSplit) != 6 {
		return res, fmt.Errorf("invalid data split length %d, expected 6", len(dataSplit))
	}

	// Date time, start and end of schedule
	date := dataSplit[1]
	timeSplit := splitTrim(dataSplit[2], "-")
	if len(timeSplit) != 2 {
		return res, fmt.Errorf("unexpected schedule time %s", dataSplit[2])
	}
	start, err := parseDateTime(date, timeSplit[0])
	if err != nil {
		return res, fmt.Errorf("failed to parse schedule start datetime: %s", err)
	}
	end, err := parseDateTime(date, timeSplit[1])
	if err != nil {
		return res, fmt.Errorf("failed to parse schedule end datetime: %s", err)
	}
	res.Start = start
	res.End = end

	// Room list
	rooms := splitTrim(dataSplit[3], "\n")
	if len(rooms) == 1 && rooms[0] == "-" {
		res.Rooms = []string{}
	} else {
		res.Rooms = rooms
	}

	// Activity and method
	activity, err := toActivity(dataSplit[4])
	if err != nil {
		return res, err
	}
	method, err := toMethod(dataSplit[5])
	if err != nil {
		return res, err
	}
	res.Activity = activity
	res.Method = method

	return res, nil
}
