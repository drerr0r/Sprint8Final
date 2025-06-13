package main

import (
	"database/sql"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// randRange использует текущий источник случайных чисел для генерации
var randRange = rand.New(rand.NewSource(time.Now().UnixNano()))

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	// get
	storedParcel, err := store.Get(number)
	assert.NoError(t, err)
	assert.Equal(t, parcel.Client, storedParcel.Client)
	assert.Equal(t, parcel.Status, storedParcel.Status)
	assert.Equal(t, parcel.Address, storedParcel.Address)
	assert.Equal(t, parcel.CreatedAt, storedParcel.CreatedAt)
	assert.Equal(t, number, storedParcel.Number)

	// delete
	err = store.Delete(number)
	assert.NoError(t, err)

	// check if parcel is deleted
	_, err = store.Get(number)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows))
}

func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(number, newAddress)
	assert.NoError(t, err)

	// check
	updatedParcel, err := store.Get(number)
	assert.NoError(t, err)
	assert.Equal(t, newAddress, updatedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	number, err := store.Add(parcel)
	assert.NoError(t, err)
	assert.NotZero(t, number)

	// set status
	newStatus := "delivered"
	err = store.SetStatus(number, newStatus)
	assert.NoError(t, err)

	// check
	updatedParcel, err := store.Get(number)
	assert.NoError(t, err)
	assert.Equal(t, newStatus, updatedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	assert.NoError(t, err)
	defer db.Close()

	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := range parcels {
		id, err := store.Add(parcels[i])
		assert.NoError(t, err)
		assert.NotZero(t, id)

		// обновляем идентификатор добавленной посылки
		parcels[i].Number = id

		// сохраняем посылку в map
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	assert.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		originalParcel, exists := parcelMap[parcel.Number]
		assert.True(t, exists)
		assert.Equal(t, originalParcel, parcel)
	}
}
