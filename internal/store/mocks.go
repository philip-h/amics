package store

func NewMockStore() Storage {
	return Storage{
		Students:    &MockStudentStore{},
		Assignments: &MockAssignmentStore{},
		Courses:     &MockCourseStore{},
	}
}

// ============================================================================
// Mock Student Store
// ============================================================================
type MockStudentStore struct {
	CreateInvoked         bool
	GetByUsernameInvoked  bool
	GetByTeacherIdInvoked bool
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

func (m *MockStudentStore) GetByTeacherId(teacherId int) ([]*Assignment, error) {
	m.GetByTeacherIdInvoked = true
	return []*Assignment{}, nil
}

// ============================================================================
// Mock Course Store
// ============================================================================
type MockCourseStore struct {
	CreateInvoked         bool
	GetByIdInvoked        bool
	GetByTeacherIdInvoked bool
	UpdateInvoked         bool
}

func (m *MockCourseStore) Create(*Course) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockCourseStore) GetById(id int) (*Course, error) {
	m.GetByIdInvoked = true
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
	GetWithGradeByStudentIdInvoked                    bool
	GetWithSubmissionByAssignmentAndStudentIdsInvoked bool
	SubmitInvoked                                     bool
	GetByIdInvoked                                    bool
	GetByCourseIdInvoked                              bool
	UpdateInvoked                                     bool
}

func (m *MockAssignmentStore) Create(*Assignment) error {
	m.CreateInvoked = true
	return nil
}

func (m *MockAssignmentStore) GetWithGradeByStudentId(int) ([]*AssignmentWithGrade, error) {
	m.GetWithGradeByStudentIdInvoked = true
	return nil, nil

}
func (m *MockAssignmentStore) GetWithSubmissionByAssignmentAndStudentIds(int, int) (*AssignmentSubmission, error) {
	m.GetWithSubmissionByAssignmentAndStudentIdsInvoked = true
	return nil, nil

}

func (m *MockAssignmentStore) Submit(int, int, *PyFile) error {
	m.SubmitInvoked = true
	return nil

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
