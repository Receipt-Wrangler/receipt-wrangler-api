package models

type GroupSettings struct {
	BaseModel
	GroupId            uint                 `gorm:"not null;unique" json:"groupId"`
	EmailToRead        GroupSettingsEmail   `json:"emailToRead"`
	SubjectLineRegexes []SubjectLineRegex   `json:"subjectLineRegexes"`
	EmailWhiteList     []GroupSettingsEmail `json:"emailWhiteList"`
}
