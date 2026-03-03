package sawit

import (
	"fmt"
	"kano/internal/database"
	"kano/internal/database/models"
)

var db = database.GetInstance()

type Sawit models.Sawit

func GetParticipantSawit(partId uint) (Sawit, error) {
	foundSawit := models.Sawit{ParticipantId: partId}
	tx := db.
		Preload("Participant.Contact").
		Where(&foundSawit).
		FirstOrCreate(&foundSawit)

	return Sawit(foundSawit), tx.Error
}

func (s Sawit) GetName() string {
	name := s.Participant.Contact.CustomName
	if name == "" {
		name = s.Participant.Contact.PushName
	}
	if name == "" {
		name = fmt.Sprintf("[Unknown participant: %s]", s.Participant.Contact.JID.User)
	}

	return name
}

func (s Sawit) Save() error {
	mSawit := models.Sawit(s)
	tx := db.Save(&mSawit)

	return tx.Error
}

func (s *Sawit) ChangeGrowDate(growDate string) {
	if s != nil {
		s.LastGrowDate = growDate
	}
}

func (s *Sawit) AddHeight(height int) {
	if s != nil {
		s.Height += height
	}
}

func (s *Sawit) WinAttack(height uint) {
	if s == nil {
		return
	}

	s.AttackTotal += 1
	s.AttackWin += 1
	s.AttackAcquiredHeight += height
	// Don't forget to modify the sawit's height
	s.AddHeight(int(height))
}

func (s *Sawit) LoseAttack(height uint) {
	if s == nil {
		return
	}

	s.AttackTotal += 1
	s.AttackLostHeight += height
	// Don't forget to modify the sawit's height
	s.AddHeight(-int(height))
}

func (s Sawit) GetWinrate() float64 {
	return float64(s.AttackWin) / float64(s.AttackTotal)
}
