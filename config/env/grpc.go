package env

import (
	"net"
	"os"

	"github.com/pkg/errors"
)

const (
	grpcHostEnvName = "GRPC_HOST"
	grpcPortEnvName = "GRPC_PORT"
)

// GRPCConfig - интерфейс конфига для инициализации GRPC-сервера.
//
// Методы:
//   - Address() string: адрес, на котором развернут GRPC-сервер в формате "хост:порт".
type GRPCConfig interface {
	Address() string
}

// grpcConfig - структура конфига GRPC-сервера, реализующая интерфейс GRPCConfig.
type grpcConfig struct {
	host string
	port string
}

// NewGRPCConfig - Метод для создания объекта конфига GRPC-сервера, реализующего
// интерфейс GRPCConfig.
// Параметры конфига берутся из переменных окружения программы.
//
// Возвращает:
//   - GRPCConfig: созданный объект конфига GRPC-сервера.
//   - error: ошибка, если что-то пошло не так.
func NewGRPCConfig() (GRPCConfig, error) {
	host := os.Getenv(grpcHostEnvName)
	if len(host) == 0 {
		return nil, errors.New("grpc host not found")
	}

	port := os.Getenv(grpcPortEnvName)
	if len(port) == 0 {
		return nil, errors.New("grpc port not found")
	}

	return &grpcConfig{
		host: host,
		port: port,
	}, nil
}

// Address string: метод для получения адреса, на котором развернут GRPC-сервер в
// формате "хост:порт".
func (cfg *grpcConfig) Address() string {
	return net.JoinHostPort(cfg.host, cfg.port)
}
