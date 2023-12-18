package pg

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/rs/zerolog/log"
)

const EventSourcing = "eventsourcing"

//go:embed migrations/*.sql
var EventSourcingFS embed.FS

type Migrations struct {
	localFS map[string]iMigration
}

type iMigration struct {
	embed.FS
	path string
}

func NewMigrations() *Migrations {
	m := &Migrations{
		localFS: map[string]iMigration{},
	}
	m.Append(EventSourcing, "migrations", EventSourcingFS)

	return m
}

func (m *Migrations) Append(key string, path string, lfs embed.FS) {
	m.localFS[key] = iMigration{
		FS:   lfs,
		path: path,
	}
}

func (m Migrations) Keys() []string {
	keys := make([]string, 0, len(m.localFS))
	for k := range m.localFS {
		keys = append(keys, k)
	}

	return keys
}

func (m Migrations) MigrateAll(dbConnStr string) error {
	for _, key := range m.Keys() {
		migrator, err := m.Get(key, dbConnStr)
		if err != nil {
			return fmt.Errorf("failed to get migrator for %s: %w", key, err)
		}

		if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to migrate %s: %w", key, err)
		}
	}

	return nil
}

func (m *Migrations) Get(key string, dbConnStr string) (*migrate.Migrate, error) {
	im, ok := m.localFS[key]
	if !ok {
		return nil, fmt.Errorf("migration %s not found", key)
	}

	return im.getMigrate(dbConnStr)
}

func (im iMigration) getMigrate(connString string) (*migrate.Migrate, error) {
	driver, err := newMigrator(im.FS, im.path)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", driver, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return m, nil
}

// func (m *Migrations) Get(key string) (embed.FS, string, error) {
// 	im, ok := m.localFS[key]
// 	if !ok {
// 		return embed.FS{}, "", fmt.Errorf("migration %s not found", key)
// 	}

// 	return im.FS, im.path, nil
// }

type driver struct {
	PartialDriver
}

func NewMigratorFS(lfs embed.FS, path string) (source.Driver, error) {
	return newMigrator(lfs, path)
	// return newMigrator(lfs, "migrations")
}

// New returns a newMigrator Driver from io/fs#FS and a relative path.
func newMigrator(fsys fs.FS, path string) (source.Driver, error) {
	var i driver
	if err := i.Init(fsys, path); err != nil {
		return nil, fmt.Errorf("failed to init driver with path %s: %w", path, err)
	}
	return &i, nil
}

// Open is part of source.Driver interface implementation.
// Open cannot be called on the iofs passthrough driver.
func (d *driver) Open(url string) (source.Driver, error) {
	return nil, errors.New("Open() cannot be called on the iofs passthrough driver")
}

// PartialDriver is a helper service for creating new source drivers working with
// io/fs.FS instances. It implements all source.Driver interface methods
// except for Open(). New driver could embed this struct and add missing Open()
// method.
//
// To prepare PartialDriver for use Init() function.
type PartialDriver struct {
	migrations *source.Migrations
	fsys       fs.FS
	path       string
}

// Init prepares not initialized IoFS instance to read migrations from a
// io/fs#FS instance and a relative path.
func (d *PartialDriver) Init(fsys fs.FS, path string) error {
	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return err
	}

	ms := source.NewMigrations()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		log.Debug().Str("file", e.Name()).Msg("found migration")
		m, err := source.DefaultParse(e.Name())
		if err != nil {
			continue
		}
		file, err := e.Info()
		if err != nil {
			return err
		}
		if !ms.Append(m) {
			return source.ErrDuplicateMigration{
				Migration: *m,
				FileInfo:  file,
			}
		}
	}

	d.fsys = fsys
	d.path = path
	d.migrations = ms
	return nil
}

// Close is part of source.Driver interface implementation.
// Closes the file system if possible.
func (d *PartialDriver) Close() error {
	c, ok := d.fsys.(io.Closer)
	if !ok {
		return nil
	}
	return c.Close()
}

// First is part of source.Driver interface implementation.
func (d *PartialDriver) First() (version uint, err error) {
	if version, ok := d.migrations.First(); ok {
		return version, nil
	}
	return 0, &fs.PathError{
		Op:   "first",
		Path: d.path,
		Err:  fs.ErrNotExist,
	}
}

// Prev is part of source.Driver interface implementation.
func (d *PartialDriver) Prev(version uint) (prevVersion uint, err error) {
	if version, ok := d.migrations.Prev(version); ok {
		return version, nil
	}
	return 0, &fs.PathError{
		Op:   "prev for version " + strconv.FormatUint(uint64(version), 10),
		Path: d.path,
		Err:  fs.ErrNotExist,
	}
}

// Next is part of source.Driver interface implementation.
func (d *PartialDriver) Next(version uint) (nextVersion uint, err error) {
	if version, ok := d.migrations.Next(version); ok {
		return version, nil
	}
	return 0, &fs.PathError{
		Op:   "next for version " + strconv.FormatUint(uint64(version), 10),
		Path: d.path,
		Err:  fs.ErrNotExist,
	}
}

// ReadUp is part of source.Driver interface implementation.
func (d *PartialDriver) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := d.migrations.Up(version); ok {
		body, err := d.open(path.Join(d.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return body, m.Identifier, nil
	}
	return nil, "", &fs.PathError{
		Op:   "read up for version " + strconv.FormatUint(uint64(version), 10),
		Path: d.path,
		Err:  fs.ErrNotExist,
	}
}

// ReadDown is part of source.Driver interface implementation.
func (d *PartialDriver) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if m, ok := d.migrations.Down(version); ok {
		body, err := d.open(path.Join(d.path, m.Raw))
		if err != nil {
			return nil, "", err
		}
		return body, m.Identifier, nil
	}
	return nil, "", &fs.PathError{
		Op:   "read down for version " + strconv.FormatUint(uint64(version), 10),
		Path: d.path,
		Err:  fs.ErrNotExist,
	}
}

func (d *PartialDriver) open(path string) (fs.File, error) {
	f, err := d.fsys.Open(path)
	if err == nil {
		return f, nil
	}
	// Some non-standard file systems may return errors that don't include the path, that
	// makes debugging harder.
	if !errors.As(err, new(*fs.PathError)) {
		err = &fs.PathError{
			Op:   "open",
			Path: path,
			Err:  err,
		}
	}
	return nil, err
}
