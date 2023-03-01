// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package gitops

import (
	"context"
	"sync"
)

// Ensure, that localRepositoryMock does implement localRepository.
// If this is not the case, regenerate this file with moq.
var _ localRepository = &localRepositoryMock{}

// localRepositoryMock is a mock implementation of localRepository.
//
//     func TestSomethingThatUseslocalRepository(t *testing.T) {
//
//         // make and configure a mocked localRepository
//         mockedlocalRepository := &localRepositoryMock{
//             CloseFunc: func(ctx context.Context)  {
// 	               panic("mock out the Close method")
//             },
//             gitCheckoutNewBranchFunc: func() error {
// 	               panic("mock out the gitCheckoutNewBranch method")
//             },
//             gitCloneFunc: func() error {
// 	               panic("mock out the gitClone method")
//             },
//             gitCommitAndPushFunc: func(message string) error {
// 	               panic("mock out the gitCommitAndPush method")
//             },
//             localPathFunc: func() string {
// 	               panic("mock out the localPath method")
//             },
//             openPullRequestFunc: func(ctx context.Context, title string, body string) (string, error) {
// 	               panic("mock out the openPullRequest method")
//             },
//             workingDirectoryCleanFunc: func() (bool, error) {
// 	               panic("mock out the workingDirectoryClean method")
//             },
//         }
//
//         // use mockedlocalRepository in code that requires localRepository
//         // and then make assertions.
//
//     }
type localRepositoryMock struct {
	// CloseFunc mocks the Close method.
	CloseFunc func(ctx context.Context)

	// gitCheckoutNewBranchFunc mocks the gitCheckoutNewBranch method.
	gitCheckoutNewBranchFunc func() error

	// gitCloneFunc mocks the gitClone method.
	gitCloneFunc func() error

	// gitCommitAndPushFunc mocks the gitCommitAndPush method.
	gitCommitAndPushFunc func(message string) error

	// localPathFunc mocks the localPath method.
	localPathFunc func() string

	// openPullRequestFunc mocks the openPullRequest method.
	openPullRequestFunc func(ctx context.Context, title string, body string) (string, error)

	// workingDirectoryCleanFunc mocks the workingDirectoryClean method.
	workingDirectoryCleanFunc func() (bool, error)

	// calls tracks calls to the methods.
	calls struct {
		// Close holds details about calls to the Close method.
		Close []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// gitCheckoutNewBranch holds details about calls to the gitCheckoutNewBranch method.
		gitCheckoutNewBranch []struct {
		}
		// gitClone holds details about calls to the gitClone method.
		gitClone []struct {
		}
		// gitCommitAndPush holds details about calls to the gitCommitAndPush method.
		gitCommitAndPush []struct {
			// Message is the message argument value.
			Message string
		}
		// localPath holds details about calls to the localPath method.
		localPath []struct {
		}
		// openPullRequest holds details about calls to the openPullRequest method.
		openPullRequest []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Title is the title argument value.
			Title string
			// Body is the body argument value.
			Body string
		}
		// workingDirectoryClean holds details about calls to the workingDirectoryClean method.
		workingDirectoryClean []struct {
		}
	}
	lockClose                 sync.RWMutex
	lockgitCheckoutNewBranch  sync.RWMutex
	lockgitClone              sync.RWMutex
	lockgitCommitAndPush      sync.RWMutex
	locklocalPath             sync.RWMutex
	lockopenPullRequest       sync.RWMutex
	lockworkingDirectoryClean sync.RWMutex
}

// Close calls CloseFunc.
func (mock *localRepositoryMock) Close(ctx context.Context) {
	if mock.CloseFunc == nil {
		panic("localRepositoryMock.CloseFunc: method is nil but localRepository.Close was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockClose.Lock()
	mock.calls.Close = append(mock.calls.Close, callInfo)
	mock.lockClose.Unlock()
	mock.CloseFunc(ctx)
}

// CloseCalls gets all the calls that were made to Close.
// Check the length with:
//     len(mockedlocalRepository.CloseCalls())
func (mock *localRepositoryMock) CloseCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockClose.RLock()
	calls = mock.calls.Close
	mock.lockClose.RUnlock()
	return calls
}

// gitCheckoutNewBranch calls gitCheckoutNewBranchFunc.
func (mock *localRepositoryMock) gitCheckoutNewBranch() error {
	if mock.gitCheckoutNewBranchFunc == nil {
		panic("localRepositoryMock.gitCheckoutNewBranchFunc: method is nil but localRepository.gitCheckoutNewBranch was just called")
	}
	callInfo := struct {
	}{}
	mock.lockgitCheckoutNewBranch.Lock()
	mock.calls.gitCheckoutNewBranch = append(mock.calls.gitCheckoutNewBranch, callInfo)
	mock.lockgitCheckoutNewBranch.Unlock()
	return mock.gitCheckoutNewBranchFunc()
}

// gitCheckoutNewBranchCalls gets all the calls that were made to gitCheckoutNewBranch.
// Check the length with:
//     len(mockedlocalRepository.gitCheckoutNewBranchCalls())
func (mock *localRepositoryMock) gitCheckoutNewBranchCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockgitCheckoutNewBranch.RLock()
	calls = mock.calls.gitCheckoutNewBranch
	mock.lockgitCheckoutNewBranch.RUnlock()
	return calls
}

// gitClone calls gitCloneFunc.
func (mock *localRepositoryMock) gitClone() error {
	if mock.gitCloneFunc == nil {
		panic("localRepositoryMock.gitCloneFunc: method is nil but localRepository.gitClone was just called")
	}
	callInfo := struct {
	}{}
	mock.lockgitClone.Lock()
	mock.calls.gitClone = append(mock.calls.gitClone, callInfo)
	mock.lockgitClone.Unlock()
	return mock.gitCloneFunc()
}

// gitCloneCalls gets all the calls that were made to gitClone.
// Check the length with:
//     len(mockedlocalRepository.gitCloneCalls())
func (mock *localRepositoryMock) gitCloneCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockgitClone.RLock()
	calls = mock.calls.gitClone
	mock.lockgitClone.RUnlock()
	return calls
}

// gitCommitAndPush calls gitCommitAndPushFunc.
func (mock *localRepositoryMock) gitCommitAndPush(message string) error {
	if mock.gitCommitAndPushFunc == nil {
		panic("localRepositoryMock.gitCommitAndPushFunc: method is nil but localRepository.gitCommitAndPush was just called")
	}
	callInfo := struct {
		Message string
	}{
		Message: message,
	}
	mock.lockgitCommitAndPush.Lock()
	mock.calls.gitCommitAndPush = append(mock.calls.gitCommitAndPush, callInfo)
	mock.lockgitCommitAndPush.Unlock()
	return mock.gitCommitAndPushFunc(message)
}

// gitCommitAndPushCalls gets all the calls that were made to gitCommitAndPush.
// Check the length with:
//     len(mockedlocalRepository.gitCommitAndPushCalls())
func (mock *localRepositoryMock) gitCommitAndPushCalls() []struct {
	Message string
} {
	var calls []struct {
		Message string
	}
	mock.lockgitCommitAndPush.RLock()
	calls = mock.calls.gitCommitAndPush
	mock.lockgitCommitAndPush.RUnlock()
	return calls
}

// localPath calls localPathFunc.
func (mock *localRepositoryMock) localPath() string {
	if mock.localPathFunc == nil {
		panic("localRepositoryMock.localPathFunc: method is nil but localRepository.localPath was just called")
	}
	callInfo := struct {
	}{}
	mock.locklocalPath.Lock()
	mock.calls.localPath = append(mock.calls.localPath, callInfo)
	mock.locklocalPath.Unlock()
	return mock.localPathFunc()
}

// localPathCalls gets all the calls that were made to localPath.
// Check the length with:
//     len(mockedlocalRepository.localPathCalls())
func (mock *localRepositoryMock) localPathCalls() []struct {
} {
	var calls []struct {
	}
	mock.locklocalPath.RLock()
	calls = mock.calls.localPath
	mock.locklocalPath.RUnlock()
	return calls
}

// openPullRequest calls openPullRequestFunc.
func (mock *localRepositoryMock) openPullRequest(ctx context.Context, title string, body string) (string, error) {
	if mock.openPullRequestFunc == nil {
		panic("localRepositoryMock.openPullRequestFunc: method is nil but localRepository.openPullRequest was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		Title string
		Body  string
	}{
		Ctx:   ctx,
		Title: title,
		Body:  body,
	}
	mock.lockopenPullRequest.Lock()
	mock.calls.openPullRequest = append(mock.calls.openPullRequest, callInfo)
	mock.lockopenPullRequest.Unlock()
	return mock.openPullRequestFunc(ctx, title, body)
}

// openPullRequestCalls gets all the calls that were made to openPullRequest.
// Check the length with:
//     len(mockedlocalRepository.openPullRequestCalls())
func (mock *localRepositoryMock) openPullRequestCalls() []struct {
	Ctx   context.Context
	Title string
	Body  string
} {
	var calls []struct {
		Ctx   context.Context
		Title string
		Body  string
	}
	mock.lockopenPullRequest.RLock()
	calls = mock.calls.openPullRequest
	mock.lockopenPullRequest.RUnlock()
	return calls
}

// workingDirectoryClean calls workingDirectoryCleanFunc.
func (mock *localRepositoryMock) workingDirectoryClean() (bool, error) {
	if mock.workingDirectoryCleanFunc == nil {
		panic("localRepositoryMock.workingDirectoryCleanFunc: method is nil but localRepository.workingDirectoryClean was just called")
	}
	callInfo := struct {
	}{}
	mock.lockworkingDirectoryClean.Lock()
	mock.calls.workingDirectoryClean = append(mock.calls.workingDirectoryClean, callInfo)
	mock.lockworkingDirectoryClean.Unlock()
	return mock.workingDirectoryCleanFunc()
}

// workingDirectoryCleanCalls gets all the calls that were made to workingDirectoryClean.
// Check the length with:
//     len(mockedlocalRepository.workingDirectoryCleanCalls())
func (mock *localRepositoryMock) workingDirectoryCleanCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockworkingDirectoryClean.RLock()
	calls = mock.calls.workingDirectoryClean
	mock.lockworkingDirectoryClean.RUnlock()
	return calls
}