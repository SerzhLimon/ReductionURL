package service

import (
	"testing"

	"github.com/SerzhLimon/ReductionURL/internal/model"
	"github.com/stretchr/testify/mock"
)

// MockRepository - простой мок для репозитория
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Get(hash string) (string, error) {
	args := m.Called(hash)
	return args.String(0), args.Error(1)
}

// Остальные методы интерфейса (заглушки)
func (m *MockRepository) Set(url, hash string) (string, error) { return "", nil }
func (m *MockRepository) Ping() error                          { return nil }
func (m *MockRepository) SetArrayURL(req []model.SetArrayURLRequest) ([]model.SetArrayURLResponse, error) {
	return nil, nil
}
func (m *MockRepository) GetArrayURL() ([]model.GetArrayURLResponse, error) { return nil, nil }
func (m *MockRepository) Delete(hash string) error                          { return nil }

// BenchmarkGetURL - бенчмарк для метода GetURL
func BenchmarkGetURL(b *testing.B) {
	// Создаем мок репозитория
	mockRepo := new(MockRepository)

	// Создаем сервис с моком
	srv := &Service{
		repo: mockRepo,
	}

	// Тестовые данные
	testHash := "a1b2c3d4"
	expectedURL := "https://example.com"

	// Настраиваем мок
	mockRepo.On("Get", testHash).Return(expectedURL, nil)

	// Сбрасываем таймер
	b.ResetTimer()

	// Запускаем бенчмарк
	for i := 0; i < b.N; i++ {
		_, err := srv.GetURL(testHash)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
