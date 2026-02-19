package store

func NewMockStore() Storage {
	return Storage{
		Students:    &MockStudentStore{},
		Assignments: &MockAssignmentStore{},
	}
}

// ============================================================================
// Mock Student Store
// ============================================================================
type MockStudentStore struct {
	CreateInvoked        bool
	GetByUsernameInvoked bool
}

func (m *MockStudentStore) Create(student *Student) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockStudentStore) GetByUsername(username string) (*Student, error) {
	m.GetByUsernameInvoked = true
	return &Student{
		Username: "testuser",
		Password: "testpass",
	}, nil
}

// ============================================================================
// Mock Assignment Store
// ============================================================================
type MockAssignmentStore struct {
	GetByIdInvoked       bool
	GetByUsernameInvoked bool
	SubmitInvoked        bool
}

func (m *MockAssignmentStore) GetById(id int, username string) (*AssignmentSubmission, error) {
	m.GetByIdInvoked = true
	return nil, nil
}

func (m *MockAssignmentStore) GetByUsername(username string) ([]*AssignmentWithGrade, error) {
	m.GetByUsernameInvoked = true
	return []*AssignmentWithGrade{}, nil
}

func (m *MockAssignmentStore) Submit(assignmentId int, username string, file *PyFile) error {
	m.SubmitInvoked = true
	return nil
}
