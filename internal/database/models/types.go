package models

type ParticipantRole string

const (
	ParticipantRoleMember     = "member"
	ParticipantRoleAdmin      = "admin"
	ParticipantRoleSuperadmin = "superadmin"
	ParticipantRoleManager    = "manager"
	ParticipantRoleLeft       = "left"
)
