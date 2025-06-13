package main

import (
	"database/sql"
	"fmt"
)

// Структура для работы с базой данных
type ParcelStore struct {
	db *sql.DB
}

// Создание нового хранилища посылок
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Добавление новой посылки в базу данных
func (s ParcelStore) Add(p Parcel) (int, error) {
	// Выполняем SQL-запрос на вставку новой записи
	result, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("ошибка создания посылки: %w", err)
	}

	// Получаем ID добавленной записи
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("ошибка получения ID посылки: %w", err)
	}

	return int(id), nil
}

// Получение посылки по её номеру
func (s ParcelStore) Get(number int) (Parcel, error) {
	var parcel Parcel
	// Выполняем SQL-запрос на получение одной записи
	err := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", number).Scan(
		&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt)
	if err != nil {
		return Parcel{}, fmt.Errorf("ошибка получения посылки: %w", err)
	}
	return parcel, nil
}

// Получение всех посылок клиента
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения посылок клиента: %w", err)
	}
	defer rows.Close()

	var parcels []Parcel
	// Проверяем ошибки после завершения итерации
	defer func() {
		if err := rows.Err(); err != nil {
			// Если есть ошибка после завершения итерации, возвращаем её
			parcels, _ = nil, fmt.Errorf("ошибка при обработке строк: %w", err)
		}
	}()

	// Проходим по всем строкам результата
	for rows.Next() {
		var parcel Parcel
		if err := rows.Scan(&parcel.Number, &parcel.Client, &parcel.Status, &parcel.Address, &parcel.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования посылки: %w", err)
		}
		parcels = append(parcels, parcel)
	}
	return parcels, nil
}

// Обновление статуса посылки
func (s ParcelStore) SetStatus(number int, status string) error {
	// Выполняем SQL-запрос на обновление статуса
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса посылки: %w", err)
	}
	return nil
}

// Обновление адреса доставки посылки
func (s ParcelStore) SetAddress(number int, address string) error {
	// Выполняем UPDATE с проверкой статуса в одном запросе
	result, err := s.db.Exec("UPDATE parcel SET address = ? WHERE number = ? AND status = ?",
		address, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("ошибка обновления адреса посылки: %w", err)
	}

	// Проверяем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	// Если строк не затронуто, значит статус не 'registered'
	if rowsAffected == 0 {
		return fmt.Errorf("изменить адрес можно только для посылок со статусом 'registered'")
	}

	return nil
}

// Удаление посылки
func (s ParcelStore) Delete(number int) error {
	// Выполняем DELETE с проверкой статуса в одном запросе
	result, err := s.db.Exec("DELETE FROM parcel WHERE number = ? AND status = ?",
		number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("ошибка удаления посылки: %w", err)
	}

	// Проверяем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества удаленных строк: %w", err)
	}

	// Если строк не затронуто, значит статус не 'registered'
	if rowsAffected == 0 {
		return fmt.Errorf("удалить посылку можно только со статусом 'registered'")
	}

	return nil
}
