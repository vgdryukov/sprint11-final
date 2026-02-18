package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

// createTable создаёт таблицу parcel, если она не существует
func createTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS parcel (
		number INTEGER PRIMARY KEY AUTOINCREMENT,
		client INTEGER NOT NULL,
		status TEXT NOT NULL,
		address TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`

	_, err := db.Exec(createTableSQL)
	return err
}

// clearTable очищает таблицу для изоляции тестов
func clearTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM parcel")
	return err
}

// setupTestDB создаёт подключение к БД и создаёт таблицу
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite", "file:test.db?cache=shared&mode=memory")
	require.NoError(t, err)

	// Создаём таблицу
	err = createTable(db)
	require.NoError(t, err)

	// Очищаем таблицу перед тестом
	err = clearTable(db)
	require.NoError(t, err)

	// Возвращаем функцию для очистки и закрытия
	cleanup := func() {
		clearTable(db) // Игнорируем ошибку при очистке
		db.Close()
	}

	return db, cleanup
}

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	var err error
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// get
	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, parcel, stored)

	// delete
	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	// Проверяем, что посылка удалена
	_, err = store.Get(parcel.Number)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestSetAddress(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	var err error
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	// check
	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, newAddress, stored.Address)
}

func TestSetStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	var err error
	parcel.Number, err = store.Add(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)

	// set status
	err = store.SetStatus(parcel.Number, ParcelStatusSent)
	require.NoError(t, err)

	// check
	stored, err := store.Get(parcel.Number)
	require.NoError(t, err)
	require.Equal(t, ParcelStatusSent, stored.Status)
}

func TestGetByClient(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := make(map[int]Parcel)

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.NotEmpty(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		expectedParcel, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		require.Equal(t, expectedParcel, parcel)
	}
}
