package messageutils

import (
	"database/sql"
	"time"
)

type Group struct {
	ID                            int64
	AccountID                     int64
	EntityID                      int64
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
	OwnerJID                      string
	Name                          string
	NameSetAt                     time.Time
	NameSetBy                     string
	Topic                         string
	TopicID                       string
	TopicSetAt                    time.Time
	TopicSetBy                    string
	TopicDeleted                  bool
	IsLocked                      bool
	IsAnnounce                    bool
	AnnonceVersionID              string
	IsEphemeral                   bool
	DisappearingTimer             int32
	IsIncognito                   bool
	IsParent                      bool
	DefaultMembershipApprovalMode string
	LinkedParentJID               sql.NullString
	IsDefaultSubgroup             bool
	IsJoinApprovalRequired        bool
	GroupCreated                  time.Time
	ParticipantVersionID          string
	MemberAddMode                 string
}
