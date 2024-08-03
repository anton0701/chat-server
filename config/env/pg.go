package env

import (
	"errors"
	"os"
)

const (
	dsnEnvName = "PG_DSN"
)

// PGConfig - интерфейс конфига для подключения к БД Postgres.
//
// Методы:
//   - DSN() string: Data Source Name - строка, содержащая всю информацию, необходимую для подключения к базе данных.
type PGConfig interface {
	DSN() string
}

// pgConfig - структура конфига для подключения к БД Postgres, реализующая интерфейс PGConfig.
type pgConfig struct {
	dsn string
}

// NewPGConfig - метод создания конфига БД Postgres, реализующего интерфейс PGConfig.
// Параметры конфига берутся из переменных окружения программы.
//
// Возвращает:
//   - PGConfig: структура конфига БД Postgres, реализующая интерфейс PGConfig.
//   - error: ошибка, если что-то пошло не так.
func NewPGConfig() (PGConfig, error) {
	dsn := os.Getenv(dsnEnvName)
	if len(dsn) == 0 {
		return nil, errors.New("pg dsn not found")
	}

	return &pgConfig{
		dsn: dsn,
	}, nil
}

// DSN - возвращает DSN для подключения к БД Postgres.
func (cfg *pgConfig) DSN() string {
	return cfg.dsn
}
