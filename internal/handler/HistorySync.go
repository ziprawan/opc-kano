package handler

import (
	"context"
	"kano/internal/config"
	"kano/internal/database"
	"kano/internal/database/models"
	"kano/internal/logger"
	"kano/internal/utils/chatutil/grouputil"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waHistorySync"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"gorm.io/gorm/clause"
)

func pushnames(l *logger.Logger, cli *whatsmeow.Client, pushes []*waHistorySync.Pushname) error {
	pushnames := map[types.JID]string{}
	pns := []types.JID{}

	for _, push := range pushes {
		if push == nil {
			l.Warnf("Pushname struct is nil, skipping")
			continue
		}

		id := push.GetID()
		name := push.GetPushname()
		l.Debugf("Got ID %s and pushname %s", id, name)
		if name == "" || id == "" {
			l.Warnf("Name or id is empty, skipping")
			continue
		}

		jid, err := types.ParseJID(id)
		if err != nil {
			l.Warnf("ID is not parsable into JID: %s", err.Error())
			continue
		}

		switch jid.Server {
		case types.DefaultUserServer:
			pns = append(pns, jid)
		case types.HiddenUserServer:
			// Do nothing
		default:
			l.Warnf("Expected JID server @lid or @s.whatsapp.net, got @%s", jid.Server)
			continue
		}

		pushnames[jid] = name
	}

	lids, err := cli.Store.LIDs.GetManyLIDsForPNs(context.Background(), pns)
	if err != nil {
		l.Errorf("Failed to get lids")
		return err
	}

	for jid, name := range pushnames {
		delete(pushnames, jid)
		lid, ok := lids[jid]
		if ok {
			pushnames[lid] = name
		}
	}

	contacts := make([]*models.Contact, len(pushnames))

	i := 0
	for jid, name := range pushnames {
		contacts[i] = &models.Contact{
			JID:      jid,
			PushName: name,
		}
		i++
	}

	db := database.GetInstance()
	tx := db.Clauses(clause.OnConflict{
		OnConstraint: "contact_jid_unique",
		DoUpdates:    clause.AssignmentColumns([]string{"push_name"}),
	}).Create(contacts)
	if tx.Error != nil {
		l.Errorf("Contacts creation returns error")
		return tx.Error
	}
	l.Debugf("Inserted %d row(s) into contact table", tx.RowsAffected)

	return nil
}

func conversations(l *logger.Logger, cli *whatsmeow.Client, convs []*waHistorySync.Conversation) error {
	for i, conv := range convs {
		if conv == nil {
			l.Warnf("Conv at idx %d is nil, skipping", i)
			continue
		}

		id := conv.GetID()
		l.Debugf("Got id %s", id)
		jid, err := types.ParseJID(id)
		if err != nil {
			l.Warnf("JID at conv idx %d is not parsable, skipping", i)
			continue
		}

		if jid.Server != types.GroupServer {
			l.Debugf("Not a group server, skipping")
			continue
		}

		grp, err := grouputil.Init(cli, jid)
		if err != nil {
			l.Errorf("Failed to init group: %s", err.Error())
			return err
		}
		l.Debugf("Group ID: %d", grp.ID)
	}

	return nil
}

func HistorySync(cli *whatsmeow.Client, ev *events.HistorySync) (err error) {
	l := config.GetLogger().Sub("HistorySync")
	l.Debugf("Got sync type: %s", ev.Data.GetSyncType().String())

	if p := ev.Data.GetPushnames(); len(p) > 0 {
		l.Debugf("Found %d pushname data(s), processing", len(p))
		err = pushnames(l.Sub("Pushname"), cli, p)
		if err != nil {
			l.Errorf("%s", err.Error())
		}
	}

	if c := ev.Data.GetConversations(); len(c) > 0 {
		l.Debugf("Found %d conversation(s), processing", len(c))
		err = conversations(l.Sub("Conversation"), cli, c)
		if err != nil {
			l.Errorf("%s", err.Error())
		}
	}

	return nil
}
