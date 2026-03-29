package avokadodb_test

import (
	"testing"

	"github.com/bilustek/avokado/avokadodb"
	"github.com/google/uuid"
)

// testModel is a test-only model embedding BaseModel.
type testModel struct {
	avokadodb.BaseModel
	Name string `gorm:"not null" json:"name"`
}

func (testModel) TableName() string { return "test_model" }

func (m testModel) GetPublicID() uuid.UUID {
	return m.BaseModel.UID
}

func TestDBModelizer_Satisfaction(t *testing.T) {
	t.Parallel()

	var _ avokadodb.DBModelizer = (*testModel)(nil)
}
