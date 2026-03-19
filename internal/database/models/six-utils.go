package models

import "go.mau.fi/whatsmeow/types"

type ClassReminder struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	Jid            types.JID `gorm:"not null;type:text;uniqueIndex:classReminder_jid_subjectClassId_offset_unique"`
	SubjectClassID uint      `gorm:"not null;uniqueIndex:classReminder_jid_subjectClassId_offset_unique"`
	AnchorAtEnd    bool      `gorm:"not null;default:false;uniqueIndex:classReminder_jid_subjectClassId_offset_unique"`
	OffsetMinutes  int       `gorm:"not null;default:0;uniqueIndex:classReminder_jid_subjectClassId_offset_unique"`

	SubjectClass *SubjectClass `gorm:"foreignKey:SubjectClassID;references:ID"`
}

func (_ ClassReminder) TableName() string {
	return "class_reminder"
}

type ClassReminderView struct {
	ScheduleId    uint      `gorm:"->"`
	ClassId       uint      `gorm:"->"`
	Jid           types.JID `gorm:"->"`
	AnchorAtEnd   bool      `gorm:"->"`
	AlertTimeUnix int64     `gorm:"->"`

	Delivery     *ClassReminderDelivery `gorm:"foreignKey:ScheduleId,Jid,AlertTimeUnix;references:ScheduleId,Jid,DeliveredForUnix;"`
	SubjectClass *SubjectClass          `gorm:"foreignKey:ClassId;references:Id"`
	Schedule     *ClassSchedule         `gorm:"foreignKey:ScheduleId;references:Id"`
}

func (_ ClassReminderView) TableName() string {
	return "class_reminder_view"
}

type ClassReminderDelivery struct {
	ScheduleId       uint      `gorm:"primaryKey"`
	Jid              types.JID `gorm:"primaryKey"`
	DeliveredForUnix int64     `gorm:"primaryKey"`
	DeliveredAt      int64     `gorm:"autoCreateTime"`
}

func (_ ClassReminderDelivery) TableName() string {
	return "class_reminder_delivery"
}

type ClassFollower struct {
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	Jid            types.JID `gorm:"not null;type:text;uniqueIndex:classReminder_jid_subjectClassId_offset_unique"`
	SubjectClassID uint      `gorm:"not null;uniqueIndex:classReminder_jid_subjectClassId_offset_unique"`

	Subject Subject `gorm:"foreignKey:SubjectClassID;references:ID"`
}

func (_ ClassFollower) TableName() string {
	return "class_follower"
}
