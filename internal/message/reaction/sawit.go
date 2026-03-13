package reaction

import (
	"errors"
	"fmt"
	"kano/internal/database/models"
	"kano/internal/message/handles/sawit"
	"kano/internal/utils/messageutil"
	"math/rand"
	"time"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

func SawitAcceptChallenge(c *messageutil.MessageContext) error {
	if c.Group != nil && !c.Group.GroupSettings.IsGameAllowed {
		c.Logger.Debugf("Game is not allowed in %s", c.Group.JID)
		return nil
	}

	c.Logger.Debugf("Entered SawitAcceptChallenge")
	if c.GetReaction() == "" {
		c.Logger.Debugf("Reaction is empty, ignoring")
		return nil
	}

	acceptorParticipantId, err := c.GetParticipantID()
	if err != nil {
		c.Logger.Errorf("%s", err)
		return err
	}

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))

	reactKey := c.GetReactionKey()
	part := reactKey.GetParticipant()
	if part == "" {
		part = reactKey.GetRemoteJID()
	}

	reactedMsgJid, _ := types.ParseJID(part)
	if !reactKey.GetFromMe() && !c.IsMe(reactedMsgJid) {
		c.Logger.Debugf("Reacted message is not from me")
		return nil
	}

	reactedId := reactKey.GetID()
	sawitAttack := models.SawitAttack{GroupId: c.Group.ID, MessageId: reactedId}
	tx := db.Where(&sawitAttack).First(&sawitAttack)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			c.Logger.Errorf("No sawit attack at this point")
			return nil
		}

		c.Logger.Errorf("Failed to get sawit attack info: %s", tx.Error)
		return tx.Error
	}

	if sawitAttack.ParticipantId == acceptorParticipantId {
		// Acceptor cannot be same as the challenger
		c.Logger.Debugf("Acceptor is same as the challenger, skipping")
		return nil
	}

	if sawitAttack.IsAttackerWin.Valid {
		// Challenge already accepted
		c.Logger.Debugf("Challenge is already accepted")
		return nil
	}

	challengerSawit, err := sawit.GetParticipantSawit(sawitAttack.ParticipantId)
	if err != nil {
		c.Logger.Errorf("Failed to get challenger sawit: %s", err)
		return err
	}
	acceptorSawit, err := sawit.GetParticipantSawit(acceptorParticipantId)
	if err != nil {
		c.Logger.Errorf("Failed to get challenger sawit: %s", err)
		return err
	}

	if acceptorSawit.Height <= 0 {
		msg := fmt.Sprintf("Dear, %s, your sawit height is negative, go pay your debt buddy 😭🙏", acceptorSawit.GetName())
		if acceptorSawit.Height == 0 {
			msg = fmt.Sprintf("Dear, %s, you don't have any sawit right now", acceptorSawit.GetName())
		}
		c.SendMessage(&waE2E.Message{
			Conversation: proto.String(msg),
		})
		return nil
	}

	isChallengerWin := r.Float32() <= 0.5

	sawitAttack.AcceptedBy.Valid = true
	sawitAttack.AcceptedBy.Int32 = int32(acceptorParticipantId)
	sawitAttack.IsAttackerWin.Valid = true
	sawitAttack.IsAttackerWin.Bool = isChallengerWin
	tx = db.Save(&sawitAttack)
	if tx.Error != nil {
		c.Logger.Errorf("Failed to save sawit attack info: %s", err)
		return err
	}

	// Save participant sawit states
	if isChallengerWin {
		// Increase the challenger (who sent the attack) height
		challengerSawit.WinAttack(sawitAttack.AttackSize)
		// And decrease the acceptor (who sent the react message) height
		acceptorSawit.LoseAttack(sawitAttack.AttackSize)
	} else {
		// Increase the acceptor (who sent the react message) height
		acceptorSawit.WinAttack(sawitAttack.AttackSize)
		// And decrease the challenger (who sent the attack) height
		challengerSawit.LoseAttack(sawitAttack.AttackSize)
	}
	err = challengerSawit.Save()
	if err != nil {
		c.Logger.Errorf("Failed to save challenger sawit info: %s", err)
		return err
	}
	err = acceptorSawit.Save()
	if err != nil {
		c.Logger.Errorf("Failed to save acceptor sawit info: %s", err)
		return err
	}

	challengerPosition, err := sawit.GetParticipantPosition(c.Group.ID, challengerSawit.ParticipantId)
	if err != nil {
		c.Logger.Errorf("Failed to get challenger sawit rank position: %s", err)
		return err
	}
	acceptorPosition, err := sawit.GetParticipantPosition(c.Group.ID, acceptorSawit.ParticipantId)
	if err != nil {
		c.Logger.Errorf("Failed to get acceptor sawit rank position: %s", err)
		return err
	}

	winnerName := challengerSawit.GetName()
	winnerWinrate := challengerSawit.GetWinrate() * 100
	winnerHeight := challengerSawit.Height
	winnerPosition := challengerPosition

	loserName := acceptorSawit.GetName()
	loserWinrate := acceptorSawit.GetWinrate() * 100
	loserHeight := acceptorSawit.Height
	loserPosition := acceptorPosition

	if !isChallengerWin {
		winnerName, loserName = loserName, winnerName
		winnerWinrate, loserWinrate = loserWinrate, winnerWinrate
		winnerHeight, loserHeight = loserHeight, winnerHeight
		winnerPosition, loserPosition = loserPosition, winnerPosition
	}

	msg := fmt.Sprintf(
		"The winner is *%s*! His sawit is now *%d cm* long. The loser's one is *%d cm*\n\n*%s*'s position in the top is *%d*.\n*%s*'s position in the top is *%d*.\n\nWin rate of the *winner — %.02f%%*\nWin rate of the *loser — %.02f%%*",
		winnerName, winnerHeight, loserHeight, winnerName, winnerPosition, loserName, loserPosition, winnerWinrate, loserWinrate,
	)
	waMsg := &waE2E.Message{Conversation: &msg}

	if time.Since(sawitAttack.CreatedAt) > (1 * time.Hour) {
		c.SendMessage(waMsg)
	} else {
		c.EditMessageWithID(sawitAttack.MessageId, waMsg)
	}

	return nil
}
