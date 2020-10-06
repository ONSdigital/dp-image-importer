// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/event"
	"sync"
)

// Ensure, that VaultClientMock does implement event.VaultClient.
// If this is not the case, regenerate this file with moq.
var _ event.VaultClient = &VaultClientMock{}

// VaultClientMock is a mock implementation of event.VaultClient.
//
//     func TestSomethingThatUsesVaultClient(t *testing.T) {
//
//         // make and configure a mocked event.VaultClient
//         mockedVaultClient := &VaultClientMock{
//             CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
// 	               panic("mock out the Checker method")
//             },
//             ReadKeyFunc: func(path string, key string) (string, error) {
// 	               panic("mock out the ReadKey method")
//             },
//         }
//
//         // use mockedVaultClient in code that requires event.VaultClient
//         // and then make assertions.
//
//     }
type VaultClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// ReadKeyFunc mocks the ReadKey method.
	ReadKeyFunc func(path string, key string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// ReadKey holds details about calls to the ReadKey method.
		ReadKey []struct {
			// Path is the path argument value.
			Path string
			// Key is the key argument value.
			Key string
		}
	}
	lockChecker sync.RWMutex
	lockReadKey sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *VaultClientMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("VaultClientMock.CheckerFunc: method is nil but VaultClient.Checker was just called")
	}
	callInfo := struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}{
		Ctx:   ctx,
		State: state,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(ctx, state)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedVaultClient.CheckerCalls())
func (mock *VaultClientMock) CheckerCalls() []struct {
	Ctx   context.Context
	State *healthcheck.CheckState
} {
	var calls []struct {
		Ctx   context.Context
		State *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// ReadKey calls ReadKeyFunc.
func (mock *VaultClientMock) ReadKey(path string, key string) (string, error) {
	if mock.ReadKeyFunc == nil {
		panic("VaultClientMock.ReadKeyFunc: method is nil but VaultClient.ReadKey was just called")
	}
	callInfo := struct {
		Path string
		Key  string
	}{
		Path: path,
		Key:  key,
	}
	mock.lockReadKey.Lock()
	mock.calls.ReadKey = append(mock.calls.ReadKey, callInfo)
	mock.lockReadKey.Unlock()
	return mock.ReadKeyFunc(path, key)
}

// ReadKeyCalls gets all the calls that were made to ReadKey.
// Check the length with:
//     len(mockedVaultClient.ReadKeyCalls())
func (mock *VaultClientMock) ReadKeyCalls() []struct {
	Path string
	Key  string
} {
	var calls []struct {
		Path string
		Key  string
	}
	mock.lockReadKey.RLock()
	calls = mock.calls.ReadKey
	mock.lockReadKey.RUnlock()
	return calls
}
