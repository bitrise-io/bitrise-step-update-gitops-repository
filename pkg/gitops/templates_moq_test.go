// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package gitops

import (
	"sync"
)

// Ensure, that AllFilesRendererMock does implement AllFilesRenderer.
// If this is not the case, regenerate this file with moq.
var _ AllFilesRenderer = &AllFilesRendererMock{}

// AllFilesRendererMock is a mock implementation of AllFilesRenderer.
//
//	func TestSomethingThatUsesAllFilesRenderer(t *testing.T) {
//
//		// make and configure a mocked AllFilesRenderer
//		mockedAllFilesRenderer := &AllFilesRendererMock{
//			renderAllFilesFunc: func() error {
//				panic("mock out the renderAllFiles method")
//			},
//		}
//
//		// use mockedAllFilesRenderer in code that requires AllFilesRenderer
//		// and then make assertions.
//
//	}
type AllFilesRendererMock struct {
	// renderAllFilesFunc mocks the renderAllFiles method.
	renderAllFilesFunc func() error

	// calls tracks calls to the methods.
	calls struct {
		// renderAllFiles holds details about calls to the renderAllFiles method.
		renderAllFiles []struct {
		}
	}
	lockrenderAllFiles sync.RWMutex
}

// renderAllFiles calls renderAllFilesFunc.
func (mock *AllFilesRendererMock) renderAllFiles() error {
	if mock.renderAllFilesFunc == nil {
		panic("AllFilesRendererMock.renderAllFilesFunc: method is nil but AllFilesRenderer.renderAllFiles was just called")
	}
	callInfo := struct {
	}{}
	mock.lockrenderAllFiles.Lock()
	mock.calls.renderAllFiles = append(mock.calls.renderAllFiles, callInfo)
	mock.lockrenderAllFiles.Unlock()
	return mock.renderAllFilesFunc()
}

// renderAllFilesCalls gets all the calls that were made to renderAllFiles.
// Check the length with:
//
//	len(mockedAllFilesRenderer.renderAllFilesCalls())
func (mock *AllFilesRendererMock) renderAllFilesCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockrenderAllFiles.RLock()
	calls = mock.calls.renderAllFiles
	mock.lockrenderAllFiles.RUnlock()
	return calls
}
