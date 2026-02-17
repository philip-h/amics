package store

func NewMockStore() Storage {
	return Storage{
		Users: &MockUserStore{},
	}
}


// ============================================================================
// Mock User Store
// ============================================================================
type MockUserStore struct {
	CreateInvoked bool
	GetByUsernameInvoked bool
}

func (m *MockUserStore) Create(user *User) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockUserStore) GetByUsername(username string) (*User, error) {
	m.GetByUsernameInvoked = true
	return &User{
		Username: "testuser",
		Password: "testpass",
	}, nil
}