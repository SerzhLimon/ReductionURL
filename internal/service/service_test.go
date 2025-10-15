package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func newWrapService() *Service {
	r := &MockRepo{}
	return &Service{
		repo: r,
	}
}

func (m *MockRepo) Get(hash string) (string, error) {
	args := m.Called(hash)
	return args.String(0), args.Error(1)
}

func (m *MockRepo) Set(url, hash string) error {
	args := m.Called(url, hash)
	return args.Error(0)
}
func TestServiceSetURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantHash string
		wantErr  error
		repoIsOn bool

		mockHash string
	}{
		{
			name:     "success set",
			url:      "someUrl",
			wantHash: "78264b7b9514f66e",
			wantErr:  nil,
			repoIsOn: true,
			mockHash: "78264b7b9514f66e",
		},
		{
			name:     "empty upl",
			url:      "",
			wantHash: "",
			wantErr:  fmt.Errorf("incorrect url"),
			repoIsOn: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newWrapService()
			mockRepo := s.repo.(*MockRepo)

			if tt.repoIsOn {
				mockRepo.On("Set", tt.url, tt.mockHash).Return(nil)
			}

			gotHash, err := s.SetURL(tt.url)

			assert.Equal(t, tt.wantHash, gotHash)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestServiceGetURL(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		wantUrl  string
		wantErr  error
		repoIsOn bool

		mockHash string
	}{
		{
			name:     "success get",
			hash:     "78264b7b9514f66e",
			wantUrl:  "someUrl",
			wantErr:  nil,
			repoIsOn: true,
			mockHash: "78264b7b9514f66e",
		},
		{
			name:     "empty upl",
			hash:     "",
			wantUrl:  "",
			wantErr:  fmt.Errorf("incorrect id"),
			repoIsOn: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newWrapService()
			mockRepo := s.repo.(*MockRepo)

			if tt.repoIsOn {
				mockRepo.On("Get", tt.mockHash).Return(tt.wantUrl, nil)
			}

			gotUrl, err := s.GetURL(tt.hash)

			assert.Equal(t, tt.wantUrl, gotUrl)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
