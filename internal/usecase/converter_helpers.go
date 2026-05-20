package usecase

import (
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/entity"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model"
	"github.com/Matrosovdream/formidable-storage-app-golang/internal/model/converter"
)

// convertField is a thin wrapper kept inside the usecase package so it can be reused
// across field-related use cases without exposing entity types in the model package.
func convertField(f *entity.FrmField) model.FrmFieldResponse {
	return converter.ToFrmFieldResponse(f)
}
