package repository_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	pg "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
)

func TestGetPins(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	page := 1
	pageSize := 3

	expectedPins := []domain.PinData{
		{FlowID: 1, Header: "title1", Description: "description1", AuthorID: 1, IsPrivate: false, MediaURL: "media_url1"},
		{FlowID: 3, Header: "title3", Description: "description3", AuthorID: 3, IsPrivate: false, MediaURL: "media_url3"},
	}

	mock.ExpectQuery(`SELECT id, title, description, author_id, is_private, media_url FROM flow LIMIT \$1 OFFSET \$2`).
		WithArgs(pageSize, (page-1)*pageSize).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "author_id", "is_private", "media_url"}).
			AddRow(1, "title1", "description1", 1, false, "media_url1").
			AddRow(2, "title2", "description2", 2, true, "media_url2").
			AddRow(3, "title3", "description3", 3, false, "media_url3"))

	repo, err := pg.NewPGPinStorage(db, "")
	require.NoError(t, err)

	pins, err := repo.GetPins(page, pageSize)
	require.NoError(t, err)

	assert.Equal(t, expectedPins, pins)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
