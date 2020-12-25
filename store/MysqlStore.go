package store

import (
	"gorm.io/gorm"
	"sync"
	"time"
)
import "gorm.io/driver/mysql"

type MysqlStore struct {
	DSN          string
	tableName    string
	mysqlClient  *gorm.DB
	MysqlVersion uint64
	locker       sync.RWMutex
}

type FilterWord struct {
	Id   int64
	Word string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMysqlStore(dsn string, tableName string) (*MysqlStore, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return &MysqlStore{
		DSN:         dsn,
		mysqlClient: db,
		tableName:   tableName,
	}, err
}

func (s *MysqlStore) Write(word string) error {
	record := FilterWord{Word: word}
	s.mysqlClient.Table(s.tableName).Where("word", word).First(&record)

	if record.Id == 0 {
		result := s.mysqlClient.Table(s.tableName).Create(&record)

		s.locker.Lock()
		defer s.locker.Unlock()
		s.MysqlVersion++

		return result.Error
	}
	return nil
}

func (s *MysqlStore) Remove(word string) error {
	result := s.mysqlClient.Table(s.tableName).Delete("word", word)
	if result.Error != nil {
		s.locker.Lock()
		defer s.locker.Unlock()
		s.MysqlVersion++
	}
	return result.Error
}

func (s *MysqlStore) ReadAll() <-chan string {
	resultChannel := make(chan string)

	go func() {
		rows, _ := s.mysqlClient.Table(s.tableName).Rows()
		defer rows.Close()
		defer close(resultChannel)
		for rows.Next() {
			result := FilterWord{}
			s.mysqlClient.ScanRows(rows, &result)
			resultChannel <- result.Word
		}
	}()

	return resultChannel
}

func (s *MysqlStore) Version() uint64 {
	return s.MysqlVersion
}
