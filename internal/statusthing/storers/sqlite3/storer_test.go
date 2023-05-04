package sqlite3

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lusis/apithings/internal/statusthing/storers"
	"github.com/lusis/apithings/internal/statusthing/storers/dbfilters"
	"github.com/lusis/apithings/internal/statusthing/types"
	_ "modernc.org/sqlite" // sql driver
)

// we're just going to rip off the sqlite library test code here:
func TempFilename(t testing.TB) string {
	f, err := ioutil.TempFile("", "statusthing-storer-tests-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func makeTestdb(t *testing.T, option string) (*sql.DB, func(), error) {
	tempFilename := TempFilename(t)
	url := tempFilename + option

	cleanupFunc := func() {
		err := os.Remove(tempFilename)
		if err != nil {
			t.Error("temp file remove error:", err)
		}
	}

	db, err := sql.Open("sqlite", url)
	if err != nil {
		return nil, nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, cleanupFunc, err
	}
	if _, err := db.Exec(createTableStatement); err != nil {
		return nil, cleanupFunc, err
	}

	return db, cleanupFunc, nil
}

func TestImplements(t *testing.T) {
	t.Parallel()
	require.Implements(t, (*storers.StatusThingStorer)(nil), &Store{})
}

func TestCreateTable(t *testing.T) {
	t.Parallel()
	db, cleanup, err := makeTestdb(t, "")
	defer cleanup()
	require.NoError(t, err)
	require.NotNil(t, db)
	s, err := New(db, true)
	require.NoError(t, err)
	require.NotNil(t, s)
}

func TestHappyPath(t *testing.T) {
	t.Parallel()
	db, cleanup, err := makeTestdb(t, "")
	defer cleanup()
	require.NoError(t, err)
	require.NotNil(t, db)
	s, err := New(db, false)
	require.NoError(t, err)
	require.NotNil(t, s)

	ctx := context.Background()
	// make sure existing db is empty
	res, err := s.Get(ctx, t.Name())
	require.ErrorIs(t, err, types.ErrNotFound)
	require.Nil(t, res)

	originalThing := &types.StatusThing{ID: t.Name() + "_id", Description: t.Name() + "_description", Name: t.Name() + "_name", Status: types.StatusGreen}
	// add service with component and validate
	ires, err := s.Insert(ctx, originalThing)
	require.NoError(t, err)
	require.NotNil(t, ires)
	require.Equal(t, originalThing, ires)

	// change status
	ures, err := s.Update(ctx, ires.ID, dbfilters.WithStatus(types.StatusYellow))
	require.NoError(t, err)
	require.NotNil(t, ures)
	require.Equal(t, ires.ID, ures.ID)
	require.Equal(t, originalThing.Name, ures.Name)
	require.Equal(t, originalThing.Description, ures.Description)

	// delete our record
	err = s.Delete(ctx, ires.ID)
	require.NoError(t, err)

	// validate delete
	finalCheck, err := s.Get(ctx, t.Name())
	require.ErrorIs(t, err, types.ErrNotFound)
	require.Nil(t, finalCheck)
}
