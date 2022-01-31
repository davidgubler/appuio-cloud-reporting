package categories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
	"github.com/appuio/appuio-cloud-reporting/pkg/erp/entity"
	"github.com/stretchr/testify/suite"
)

type CategoriesSuite struct {
	dbtest.Suite
}

func (s *CategoriesSuite) TestReconcile() {

	s.Run("GivenCategoryWithEmptyTarget_ThenExpectUpdateAfterReconciler", func() {

		cat := db.Category{Source: "us-rac-2:disposal-plant-p-12a-furnace-control"}

		s.Require().NoError(
			db.GetNamed(s.DB(), &cat, "INSERT INTO categories (source,target) VALUES (:source,:target) RETURNING *", cat),
		)

		err := Reconcile(context.Background(), s.DB(), &stubReconciler{returnArg: entity.Category{Source: cat.Source, Target: "12"}, returnErr: nil})
		s.Require().NoError(err)

		s.Require().NoError(
			db.GetNamed(s.DB(), &cat, "SELECT * FROM categories WHERE source=:source", cat),
		)
		s.Equal("us-rac-2:disposal-plant-p-12a-furnace-control", cat.Source) // Verify unchanged
		s.Equal("12", cat.Target.String)                                     // Verify updated
		s.True(cat.Target.Valid)
	})

	s.Run("GivenCategoryWithSetTarget_ThenDoNothing", func() {
		cat := db.Category{Source: "us-rac-2:nest-elevator-control", Target: sql.NullString{String: "12", Valid: true}}

		s.Require().NoError(
			db.GetNamed(s.DB(), &cat, "INSERT INTO categories (source,target) VALUES (:source,:target) RETURNING *", cat),
		)

		err := Reconcile(context.Background(), s.DB(), &stubReconciler{returnArg: entity.Category{Source: cat.Source, Target: cat.Target.String}, returnErr: nil})
		s.Require().NoError(err)

		s.Require().NoError(
			db.GetNamed(s.DB(), &cat, "SELECT * FROM categories WHERE source=:source", cat),
		)

		s.Equal("us-rac-2:nest-elevator-control", cat.Source) // Verify unchanged
		s.Equal("12", cat.Target.String)                      // Verify unchanged
		s.True(cat.Target.Valid)
	})
}

func TestCategories(t *testing.T) {
	suite.Run(t, new(CategoriesSuite))
}

type stubReconciler struct {
	returnErr error
	returnArg entity.Category
}

func (s *stubReconciler) Reconcile(_ context.Context, _ entity.Category) (entity.Category, error) {
	return s.returnArg, s.returnErr
}
