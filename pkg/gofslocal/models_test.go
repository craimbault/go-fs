package gofslocal

import (
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

type mockFileInfo struct {
	mock.Mock
}

func (m *mockFileInfo) Name() string {
	return m.Called().String(0)
}

func (m *mockFileInfo) Size() int64 {
	return m.Called().Get(0).(int64)
}

func (m *mockFileInfo) Mode() os.FileMode {
	return m.Called().Get(0).(os.FileMode)
}

func (m *mockFileInfo) ModTime() time.Time {
	return m.Called().Get(0).(time.Time)
}

func (m *mockFileInfo) IsDir() bool {
	return m.Called().Get(0).(bool)
}

func (m *mockFileInfo) Sys() any {
	return m.Called().Get(0)
}
