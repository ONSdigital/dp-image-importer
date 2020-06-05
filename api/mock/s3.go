// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/api"
	"sync"
)

var (
	lockS3ClienterMockChecker sync.RWMutex
)

// Ensure, that S3ClienterMock does implement S3Clienter.
// If this is not the case, regenerate this file with moq.
var _ api.S3Clienter = &S3ClienterMock{}

// S3ClienterMock is a mock implementation of api.S3Clienter.
//
//     func TestSomethingThatUsesS3Clienter(t *testing.T) {
//
//         // make and configure a mocked api.S3Clienter
//         mockedS3Clienter := &S3ClienterMock{
//             CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//         }
//
//         // use mockedS3Clienter in code that requires api.S3Clienter
//         // and then make assertions.
//
//     }
type S3ClienterMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
	}
}

// Checker calls CheckerFunc.
func (mock *S3ClienterMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("S3ClienterMock.CheckerFunc: method is nil but S3Clienter.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	lockS3ClienterMockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	lockS3ClienterMockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedS3Clienter.CheckerCalls())
func (mock *S3ClienterMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	lockS3ClienterMockChecker.RLock()
	calls = mock.calls.Checker
	lockS3ClienterMockChecker.RUnlock()
	return calls
}