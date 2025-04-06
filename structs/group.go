package structs

import "database/sql"

type BasicGroupInfo struct {
	ID    int64          `json:"id"`
	Name  string         `json:"name"`
	Topic sql.NullString `json:"topic"`
}

type GroupInfo struct {
	BasicGroupInfo

	Titles     []Title `json:"titles"`
	MemberID   int64   `json:"member_id"`
	MemberRole string  `json:"member_role"`
}

const (
	MemberRoleSuperAdmin = "SUPERADMIN"
	MemberRoleAdmin      = "ADMIN"
	MemberRoleManager    = "MANAGER"
	MemberRoleMember     = "MEMBER"
)
