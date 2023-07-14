package commands

type UpdateProfileCommand struct {
	DisplayName        string `json:"displayName"`
	DefaultAvatarColor string `json:"defaultAvatarColor"`
}
