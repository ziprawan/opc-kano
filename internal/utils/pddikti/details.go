package pddikti

import (
	"encoding/json"
	"fmt"
)

func GetMHSDetails(mhsID string) (*DiddyDetailsMHS, error) {
	url := buildUrl("detail", "mhs", mhsID)
	fmt.Println("Fetching:", url)
	resp, err := fetch(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DiddyDetailsMHS
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetPNSDetails(pnsID string) (*DiddyDetailsPNS, error) {
	detail := DiddyDetailsPNS{}
	detail.TeachingHistories = map[string][]DiddyPNSTeachHistory{}

	// PNS Detail
	profileUrl := buildUrl("dosen", "profile", pnsID)
	resp, err := fetch(profileUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&detail.Profile)
	if err != nil {
		return nil, err
	}

	// Study Histories
	sHistoryUrl := buildUrl("dosen", "study-history", pnsID)
	resp, err = fetch(sHistoryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&detail.StudyHistories)
	if err != nil {
		return nil, err
	}

	// Teaching Histories
	var tHistories []DiddyPNSTeachHistory
	tHistoryUrl := buildUrl("dosen", "teaching-history", pnsID)
	resp, err = fetch(tHistoryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&tHistories)
	if err != nil {
		return nil, err
	}
	for _, t := range tHistories {
		detail.TeachingHistories[t.NamaSemester] = append(detail.TeachingHistories[t.NamaSemester], t)
	}

	portfolios := []string{"penelitian", "pengabdian", "karya", "paten"}
	for _, portfolioType := range portfolios {
		portfolioUrl := buildUrl("dosen", "portofolio", portfolioType, pnsID)
		resp, err = fetch(portfolioUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		switch portfolioType {
		case "penelitian":
			err = json.NewDecoder(resp.Body).Decode(&detail.Researches)
		case "pengabdian":
			err = json.NewDecoder(resp.Body).Decode(&detail.Devotionals)
		case "karya":
			err = json.NewDecoder(resp.Body).Decode(&detail.Creations)
		case "paten":
			err = json.NewDecoder(resp.Body).Decode(&detail.Patents)
		}
		if err != nil {
			return nil, err
		}
	}

	return &detail, nil
}
