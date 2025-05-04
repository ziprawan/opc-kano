package saveutils

import (
	"database/sql"
	"time"

	"go.mau.fi/whatsmeow/types"
)

type GroupSettings struct {
	ChosenShipping   []string
	LastShippingTime sql.NullTime
}

type Group struct {
	ID                            int64
	JID                           *types.JID
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

	Settings *GroupSettings
}

type ContactSettings struct {
	ConfessTargetID   sql.NullInt64
	CurrentWordle     sql.NullString
	WordleGeneratedAt sql.NullTime
	WordleGuesses     []string
}

type Contact struct {
	ID                  int64
	EntityID            int64
	AccountID           int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
	JID                 *types.JID
	CustomName          sql.NullString
	PushName            sql.NullString
	LoginRequestID      sql.NullString
	LoginExpirationDate sql.NullTime
	LoginRedirect       sql.NullString

	Settings *ContactSettings
}

type ParticipantRole string

const (
	ParticipantRoleSuperAdmin ParticipantRole = "SUPERADMIN"
	ParticipantRoleAdmin      ParticipantRole = "ADMIN"
	ParticipantRoleManager    ParticipantRole = "MANAGER"
	ParticipantRoleMember     ParticipantRole = "MEMBER"
	ParticipantRoleLeft       ParticipantRole = "LEFT"
)

type Participant struct {
	ID        int64
	GroupID   int64
	ContactID int64
	Role      ParticipantRole
}
