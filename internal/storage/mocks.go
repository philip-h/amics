package storage

func NewMockStore() Storage {
	return Storage{
		People:      &MockPersonStore{},
		Assignments: &MockAssignmentStore{},
		Courses:     &MockCourseStore{},
	}
}

// ============================================================================
// Mock Student Store
// ============================================================================
type MockPersonStore struct {
	CreateInvoked         bool
	GetByIdInvoked        bool
	GetByUsernameInvoked  bool
	GetByTeacherIdInvoked bool
	GetByCourseIdInvoked  bool
	ChangePasswordInvoked bool

	CompareHashAndPasswordInvoked bool
}

func (m *MockPersonStore) Create(*Person, int) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockPersonStore) GetById(id int) (*Person, error) {
	m.GetByIdInvoked = true
	return &Person{
		Username: "testuser",
		Password: "testpass",
	}, nil
}

func (m *MockPersonStore) GetByUsername(username string) (*Person, error) {
	m.GetByUsernameInvoked = true
	return &Person{
		Username: "testuser",
		Password: "testpass",
	}, nil
}

func (m *MockPersonStore) GetByTeacherId(teacherId int) ([]*Assignment, error) {
	m.GetByTeacherIdInvoked = true
	return []*Assignment{}, nil
}

func (m *MockPersonStore) GetByCourseId(int) ([]*Person, error) {
	m.GetByCourseIdInvoked = true
	return nil, nil
}

func (m *MockPersonStore) ChangePassword(int, string) error {
	m.ChangePasswordInvoked = true
	return nil
}

func (m *MockPersonStore) CompareHashAndPassword(string, string) bool {
	m.CompareHashAndPasswordInvoked = true
	return true
}

// ============================================================================
// Mock Course Store
// ============================================================================
type MockCourseStore struct {
	CreateInvoked         bool
	GetByIdInvoked        bool
	GetByJoinCodeInvoked  bool
	GetByTeacherIdInvoked bool
	UpdateInvoked         bool
}

func (m *MockCourseStore) Create(*Course, int) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockCourseStore) GetById(id int) (*Course, error) {
	m.GetByIdInvoked = true
	return &Course{}, nil
}
func (m *MockCourseStore) GetByJoinCode(string) (*Course, error) {
	m.GetByJoinCodeInvoked = true
	return &Course{}, nil
}
func (m *MockCourseStore) GetByTeacherId(id int) ([]*Course, error) {
	m.GetByTeacherIdInvoked = true
	return []*Course{}, nil
}

func (m *MockCourseStore) Update(*Course) error {
	m.UpdateInvoked = true
	return nil
}

// ============================================================================
// Mock Assignment Store
// ============================================================================
type MockAssignmentStore struct {
	CreateInvoked                                     bool
	GetWithGradeInvoked                               bool
	GetWithSubmissionByAssignmentAndStudentIdsInvoked bool
	SubmitInvoked                                     bool
	GetByIdInvoked                                    bool
	GetByCourseIdInvoked                              bool
	UpdateInvoked                                     bool
	UpdateAllInvoked                                  bool
}

func (m *MockAssignmentStore) Create(*Assignment) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockAssignmentStore) GetWithGrade(int) ([]*AssignmentWithGrade, error) {
	m.GetWithGradeInvoked = true
	return nil, nil

}
func (m *MockAssignmentStore) GetWithSubmissionByAssignmentAndStudentIds(int, int) (*AssignmentSubmission, error) {
	m.GetWithSubmissionByAssignmentAndStudentIdsInvoked = true
	return nil, nil

}

func (m *MockAssignmentStore) GetById(int) (*Assignment, error) {
	m.GetByIdInvoked = true
	return nil, nil

}
func (m *MockAssignmentStore) GetByCourseId(int) ([]*Assignment, error) {
	m.GetByCourseIdInvoked = true
	return nil, nil
}

func (m *MockAssignmentStore) Update(*Assignment) error {
	m.UpdateInvoked = true
	return nil
}

func (m *MockAssignmentStore) UpdateAll(int, [][]string) error {
	m.UpdateAllInvoked = true
	return nil
}
