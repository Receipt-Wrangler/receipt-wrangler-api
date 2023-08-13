package commands

import "receipt-wrangler/api/internal/models"

type QuickScanCommand struct {
 ImageData []byte `json:"imageData"`
 GroupId uint `json:"groupId"`
 Group models.Group `json:"-"`
 Status models.GroupStatus `json:"groupStatus"`
}