package sqlite3

// TODO: consider migrating to proper sql builder if complexity grows
// TODO: centralize transaction error handling to ensure consistency
import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lusis/apithings/internal/statusthing/storers/dbfilters"
	"github.com/lusis/apithings/internal/statusthing/types"

	"modernc.org/sqlite"
	_ "modernc.org/sqlite" // sql driver
)

const (
	thingTableName = "statusthings"
)

var (
	selectStatement      = fmt.Sprintf("SELECT id,name,description,status from %s where id = ?", thingTableName)
	selectAllStatement   = fmt.Sprintf("SELECT id,name,description,status from %s", thingTableName)
	insertStatement      = fmt.Sprintf("INSERT INTO %s (id, name, description, status) VALUES (?,?,?,?)", thingTableName)
	deleteStatement      = fmt.Sprintf("DELETE FROM %s where id = ?", thingTableName)
	createTableStatement = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (`id` VARCHAR(191) PRIMARY KEY, `name` VARCHAR(191) NOT NULL UNIQUE, `description` VARCHAR(191) DEFAULT NULL, `status` INT UNSIGNED NOT NULL)", thingTableName)
)

// Store is something that can store [types.StatusThing]
type Store struct {
	db *sql.DB
}

// statusThingRecord is the sqlite representation of a [types.StatusThing]
type statusThingRecord struct {
	id          string
	name        string
	description string
	status      int
}

// converts from db representation
func (s *statusThingRecord) toStatusThing() (*types.StatusThing, error) {
	st := &types.StatusThing{
		ID:          s.id,
		Name:        s.name,
		Description: s.description,
		Status:      types.Status(s.status),
	}

	return st, nil
}

// converts to db representation
func toRecord(st *types.StatusThing) (*statusThingRecord, error) { // nolint: unparam
	res := &statusThingRecord{
		id:          st.ID,
		description: st.Description,
		name:        st.Name,
		status:      int(st.Status),
	}

	return res, nil
}

// New returns a new sqlite3-backed service storer
func New(db *sql.DB, createTable bool) (*Store, error) {
	if createTable {
		if _, err := db.ExecContext(context.TODO(), createTableStatement); err != nil {
			return nil, fmt.Errorf("unable to create table: %w", err)
		}
	}
	return &Store{db: db}, nil
}

// Get gets a thing
func (ss *Store) Get(ctx context.Context, id string) (*types.StatusThing, error) {
	st := &statusThingRecord{}
	if err := ss.db.QueryRowContext(ctx, selectStatement, id).Scan(&st.id, &st.name, &st.description, &st.status); err != nil {
		if err == sql.ErrNoRows {
			return nil, types.ErrNotFound
		}
		return nil, fmt.Errorf("unable to query for services: %w", err)
	}

	return st.toStatusThing()
}

// GetAll gets all records from the store
func (ss *Store) GetAll(ctx context.Context) ([]*types.StatusThing, error) {
	res := []*types.StatusThing{}
	rows, err := ss.db.QueryContext(ctx, selectAllStatement)
	if err == sql.ErrNoRows {
		return res, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to query: %w", err)
	}
	for rows.Next() {
		rec := &statusThingRecord{}
		if err := rows.Scan(&rec.id, &rec.name, &rec.description, &rec.status); err != nil {
			return nil, fmt.Errorf("unable to read data: %w", err)
		}
		r, err := rec.toStatusThing()
		if err != nil {
			return nil, fmt.Errorf("unable to transform: %w", err)
		}
		res = append(res, r)
	}
	return res, nil
}

// Insert adds a thing to the db
func (ss *Store) Insert(ctx context.Context, thing *types.StatusThing) (*types.StatusThing, error) {
	st, err := toRecord(thing)
	if err != nil {
		return nil, err
	}

	tx, err := ss.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}

	rows, err := tx.ExecContext(ctx, insertStatement, st.id, st.name, st.description, st.status)
	var sqliteError = &sqlite.Error{}
	if errors.As(err, &sqliteError) && sqliteError.Code() == 2067 {
		return nil, types.ErrAlreadyExists
	}
	if err != nil {
		return nil, ss.rollback(tx, err)
	}

	if _, err := rows.RowsAffected(); err != nil {
		return nil, ss.rollback(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return ss.Get(ctx, st.id)
}

// Update updates a thing
func (ss *Store) Update(ctx context.Context, id string, opts ...dbfilters.Option) (*types.StatusThing, error) {
	dbopts, err := dbfilters.New(opts...)
	if err != nil {
		return nil, err
	}

	if _, err := ss.Get(ctx, id); err != nil {
		return nil, err
	}
	tx, err := ss.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	// UnknownValue is not the zero-value for types.Status
	if dbopts.Status() != types.StatusUnknown {
		stmt := fmt.Sprintf("UPDATE %s SET status = ?", thingTableName)
		if _, err := tx.ExecContext(ctx, stmt, dbopts.Status()); err != nil {
			return nil, ss.rollback(tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to save data: %w", err)
	}
	return ss.Get(ctx, id)
}

// Delete removes a thing from the db
func (ss *Store) Delete(ctx context.Context, id string) error {
	tx, err := ss.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, deleteStatement, id)
	if err != nil && err != sql.ErrNoRows {
		return ss.rollback(tx, err)
	}
	if err == sql.ErrNoRows {
		// no rows are fine
		return nil
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return ss.rollback(tx, err)
	}
	if affected != 1 {
		return ss.rollback(tx, fmt.Errorf("more than one row deleted. this should not happen"))
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("unable to save data: %w", err)
	}
	return nil
}

// codify rollback behaviour centrally
func (ss *Store) rollback(tx *sql.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("unable to rollback transaction for original error %w: %w", err, rollbackErr)
	}
	return err
}
