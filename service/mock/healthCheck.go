// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/service"
	"net/http"
	"sync"
)

var (
	lockHealthCheckerMockAddCheck sync.RWMutex
	lockHealthCheckerMockHandler  sync.RWMutex
	lockHealthCheckerMockStart    sync.RWMutex
	lockHealthCheckerMockStop     sync.RWMutex
)

// Ensure, that HealthCheckerMock does implement HealthChecker.
// If this is not the case, regenerate this file with moq.
var _ service.HealthChecker = &HealthCheckerMock{}

// HealthCheckerMock is a mock implementation of service.HealthChecker.
//
//     func TestSomethingThatUsesHealthChecker(t *testing.T) {
//
//         // make and configure a mocked service.HealthChecker
//         mockedHealthChecker := &HealthCheckerMock{
//             AddCheckFunc: func(name string, checker healthcheck.Checker) error {
// 	               panic("mock out the AddCheck method")
//             },
//             HandlerFunc: func(w http.ResponseWriter, req *http.Request)  {
// 	               panic("mock out the Handler method")
//             },
//             StartFunc: func(ctx context.Context)  {
// 	               panic("mock out the Start method")
//             },
//             StopFunc: func()  {
// 	               panic("mock out the Stop method")
//             },
//         }
//
//         // use mockedHealthChecker in code that requires service.HealthChecker
//         // and then make assertions.
//
//     }
type HealthCheckerMock struct {
	// AddCheckFunc mocks the AddCheck method.
	AddCheckFunc func(name string, checker healthcheck.Checker) error

	// HandlerFunc mocks the Handler method.
	HandlerFunc func(w http.ResponseWriter, req *http.Request)

	// StartFunc mocks the Start method.
	StartFunc func(ctx context.Context)

	// StopFunc mocks the Stop method.
	StopFunc func()

	// calls tracks calls to the methods.
	calls struct {
		// AddCheck holds details about calls to the AddCheck method.
		AddCheck []struct {
			// Name is the name argument value.
			Name string
			// Checker is the checker argument value.
			Checker healthcheck.Checker
		}
		// Handler holds details about calls to the Handler method.
		Handler []struct {
			// W is the w argument value.
			W http.ResponseWriter
			// Req is the req argument value.
			Req *http.Request
		}
		// Start holds details about calls to the Start method.
		Start []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Stop holds details about calls to the Stop method.
		Stop []struct {
		}
	}
}

// AddCheck calls AddCheckFunc.
func (mock *HealthCheckerMock) AddCheck(name string, checker healthcheck.Checker) error {
	if mock.AddCheckFunc == nil {
		panic("HealthCheckerMock.AddCheckFunc: method is nil but HealthChecker.AddCheck was just called")
	}
	callInfo := struct {
		Name    string
		Checker healthcheck.Checker
	}{
		Name:    name,
		Checker: checker,
	}
	lockHealthCheckerMockAddCheck.Lock()
	mock.calls.AddCheck = append(mock.calls.AddCheck, callInfo)
	lockHealthCheckerMockAddCheck.Unlock()
	return mock.AddCheckFunc(name, checker)
}

// AddCheckCalls gets all the calls that were made to AddCheck.
// Check the length with:
//     len(mockedHealthChecker.AddCheckCalls())
func (mock *HealthCheckerMock) AddCheckCalls() []struct {
	Name    string
	Checker healthcheck.Checker
} {
	var calls []struct {
		Name    string
		Checker healthcheck.Checker
	}
	lockHealthCheckerMockAddCheck.RLock()
	calls = mock.calls.AddCheck
	lockHealthCheckerMockAddCheck.RUnlock()
	return calls
}

// Handler calls HandlerFunc.
func (mock *HealthCheckerMock) Handler(w http.ResponseWriter, req *http.Request) {
	if mock.HandlerFunc == nil {
		panic("HealthCheckerMock.HandlerFunc: method is nil but HealthChecker.Handler was just called")
	}
	callInfo := struct {
		W   http.ResponseWriter
		Req *http.Request
	}{
		W:   w,
		Req: req,
	}
	lockHealthCheckerMockHandler.Lock()
	mock.calls.Handler = append(mock.calls.Handler, callInfo)
	lockHealthCheckerMockHandler.Unlock()
	mock.HandlerFunc(w, req)
}

// HandlerCalls gets all the calls that were made to Handler.
// Check the length with:
//     len(mockedHealthChecker.HandlerCalls())
func (mock *HealthCheckerMock) HandlerCalls() []struct {
	W   http.ResponseWriter
	Req *http.Request
} {
	var calls []struct {
		W   http.ResponseWriter
		Req *http.Request
	}
	lockHealthCheckerMockHandler.RLock()
	calls = mock.calls.Handler
	lockHealthCheckerMockHandler.RUnlock()
	return calls
}

// Start calls StartFunc.
func (mock *HealthCheckerMock) Start(ctx context.Context) {
	if mock.StartFunc == nil {
		panic("HealthCheckerMock.StartFunc: method is nil but HealthChecker.Start was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	lockHealthCheckerMockStart.Lock()
	mock.calls.Start = append(mock.calls.Start, callInfo)
	lockHealthCheckerMockStart.Unlock()
	mock.StartFunc(ctx)
}

// StartCalls gets all the calls that were made to Start.
// Check the length with:
//     len(mockedHealthChecker.StartCalls())
func (mock *HealthCheckerMock) StartCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	lockHealthCheckerMockStart.RLock()
	calls = mock.calls.Start
	lockHealthCheckerMockStart.RUnlock()
	return calls
}

// Stop calls StopFunc.
func (mock *HealthCheckerMock) Stop() {
	if mock.StopFunc == nil {
		panic("HealthCheckerMock.StopFunc: method is nil but HealthChecker.Stop was just called")
	}
	callInfo := struct {
	}{}
	lockHealthCheckerMockStop.Lock()
	mock.calls.Stop = append(mock.calls.Stop, callInfo)
	lockHealthCheckerMockStop.Unlock()
	mock.StopFunc()
}

// StopCalls gets all the calls that were made to Stop.
// Check the length with:
//     len(mockedHealthChecker.StopCalls())
func (mock *HealthCheckerMock) StopCalls() []struct {
} {
	var calls []struct {
	}
	lockHealthCheckerMockStop.RLock()
	calls = mock.calls.Stop
	lockHealthCheckerMockStop.RUnlock()
	return calls
}
