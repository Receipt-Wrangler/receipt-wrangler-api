package models

type FileDataView struct {
	BaseModel
	EncodedImage string `json:"encodedImage"`
	Name         string `json:"name"`
}

func (view FileDataView) FromFileData(fileData FileData) FileDataView {
	return FileDataView{
		BaseModel: BaseModel{
			ID:              fileData.ID,
			CreatedAt:       fileData.CreatedAt,
			UpdatedAt:       fileData.UpdatedAt,
			CreatedBy:       fileData.CreatedBy,
			CreatedByString: fileData.CreatedByString,
		},
		EncodedImage: "",
		Name:         fileData.Name,
	}
}
