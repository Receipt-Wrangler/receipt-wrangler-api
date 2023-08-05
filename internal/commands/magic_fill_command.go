package commands

type MagicFillCommand struct {
	ImageData []byte `json:"imageData"`
	Filename  string `json:"filename"`
}
