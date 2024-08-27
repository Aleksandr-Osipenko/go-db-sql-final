package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err) //тест на ошибку подключения к БД
	}

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавление новой посылкуи в БД
	add, err := store.Add(parcel)
	require.NoError(t, err)  //тест на ошибку добавления в БД
	require.NotEmpty(t, add) //тест на наличие идентификатора

	// get
	// получение только что добавленной посылки
	get, err := store.Get(add)
	parcel.Number = add
	assert.NoError(t, err)       //тест на ошибку получения только что добавленной посылки
	assert.Equal(t, parcel, get) //тест на совпадение полей в полученном объекте со значениями полей в переменной parcel

	// delete
	// удаление добавленной посылки
	err = store.Delete(add)
	require.NoError(t, err) //тест на отсутствие ошибки удаления посылки
	_, err = store.Get(add)
	assert.Error(t, err) //тест на наличие ошибки при получении удалённой посылки
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настройка подключения к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err) //тест на ошибку подключения к БД
	}

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавление новой посылкуи в БД
	add, err := store.Add(parcel)
	require.NoError(t, err)  //тест на ошибку добавления в БД
	require.NotEmpty(t, add) //тест на наличие идентификатора

	// set address
	// обновление адреса
	newAddress := "new test address"
	err = store.SetAddress(add, newAddress)
	require.NoError(t, err) //тест на ошибку обновления адреса

	// check
	// получение добавленной посылки
	get, err := store.Get(add)
	assert.NoError(t, err)                   //тест на ошибку получения посылки
	assert.Equal(t, newAddress, get.Address) //сравнение нового адреса посылки с заданным
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err) //тест на ошибку подключения к БД
	}

	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавление новой посылкуи в БД
	add, err := store.Add(parcel)
	require.NoError(t, err)  //тест на ошибку добавления в БД
	require.NotEmpty(t, add) //тест на наличие идентификатора

	// set status
	// обновление статуа
	newStatus := "ParcelStatusSent"
	err = store.SetStatus(add, newStatus)
	require.NoError(t, err) //тест на ошибку обновления статуса

	// check
	// получение добавленной посылки
	get, err := store.Get(add)
	assert.NoError(t, err)                 //тест на ошибку получения посылки
	assert.Equal(t, newStatus, get.Status) //сравнение нового статуса посылки с заданным
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err) //тест на ошибку подключения к БД
	}

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
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		// добавление новой посылкуи в БД
		id, err := store.Add(parcels[i])

		require.NoError(t, err) //тест на ошибку добавления в БД
		require.NotEmpty(t, id) //тест на наличие идентификатора

		// обновляем идентификатор у добавленной посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]

	}

	// get by client
	storedParcels, err := store.GetByClient(client)    // получите список посылок по идентификатору клиента, сохранённого в переменной client
	require.NoError(t, err)                            // тест на ошибку
	require.Equal(t, len(parcels), len(storedParcels)) // убедитесь, что количество полученных посылок совпадает с количеством добавленных

	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		assert.Equal(t, parcelMap[parcel.Number], parcel)
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
