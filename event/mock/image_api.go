// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/image"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-image-importer/event"
	"sync"
)

// Ensure, that ImageAPIClientMock does implement event.ImageAPIClient.
// If this is not the case, regenerate this file with moq.
var _ event.ImageAPIClient = &ImageAPIClientMock{}

// ImageAPIClientMock is a mock implementation of event.ImageAPIClient.
//
//	func TestSomethingThatUsesImageAPIClient(t *testing.T) {
//
//		// make and configure a mocked event.ImageAPIClient
//		mockedImageAPIClient := &ImageAPIClientMock{
//			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error {
//				panic("mock out the Checker method")
//			},
//			GetImageFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string) (image.Image, error) {
//				panic("mock out the GetImage method")
//			},
//			PostDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.NewImageDownload) (image.ImageDownload, error) {
//				panic("mock out the PostDownloadVariant method")
//			},
//			PutDownloadVariantFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error) {
//				panic("mock out the PutDownloadVariant method")
//			},
//			PutImageFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.Image) (image.Image, error) {
//				panic("mock out the PutImage method")
//			},
//		}
//
//		// use mockedImageAPIClient in code that requires event.ImageAPIClient
//		// and then make assertions.
//
//	}
type ImageAPIClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(ctx context.Context, state *healthcheck.CheckState) error

	// GetImageFunc mocks the GetImage method.
	GetImageFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string) (image.Image, error)

	// PostDownloadVariantFunc mocks the PostDownloadVariant method.
	PostDownloadVariantFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.NewImageDownload) (image.ImageDownload, error)

	// PutDownloadVariantFunc mocks the PutDownloadVariant method.
	PutDownloadVariantFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error)

	// PutImageFunc mocks the PutImage method.
	PutImageFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.Image) (image.Image, error)

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// State is the state argument value.
			State *healthcheck.CheckState
		}
		// GetImage holds details about calls to the GetImage method.
		GetImage []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ImageID is the imageID argument value.
			ImageID string
		}
		// PostDownloadVariant holds details about calls to the PostDownloadVariant method.
		PostDownloadVariant []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ImageID is the imageID argument value.
			ImageID string
			// Data is the data argument value.
			Data image.NewImageDownload
		}
		// PutDownloadVariant holds details about calls to the PutDownloadVariant method.
		PutDownloadVariant []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ImageID is the imageID argument value.
			ImageID string
			// Variant is the variant argument value.
			Variant string
			// Data is the data argument value.
			Data image.ImageDownload
		}
		// PutImage holds details about calls to the PutImage method.
		PutImage []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// CollectionID is the collectionID argument value.
			CollectionID string
			// ImageID is the imageID argument value.
			ImageID string
			// Data is the data argument value.
			Data image.Image
		}
	}
	lockChecker             sync.RWMutex
	lockGetImage            sync.RWMutex
	lockPostDownloadVariant sync.RWMutex
	lockPutDownloadVariant  sync.RWMutex
	lockPutImage            sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *ImageAPIClientMock) Checker(ctx context.Context, state *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("ImageAPIClientMock.CheckerFunc: method is nil but ImageAPIClient.Checker was just called")
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
//
//	len(mockedImageAPIClient.CheckerCalls())
func (mock *ImageAPIClientMock) CheckerCalls() []struct {
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

// GetImage calls GetImageFunc.
func (mock *ImageAPIClientMock) GetImage(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string) (image.Image, error) {
	if mock.GetImageFunc == nil {
		panic("ImageAPIClientMock.GetImageFunc: method is nil but ImageAPIClient.GetImage was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ImageID:          imageID,
	}
	mock.lockGetImage.Lock()
	mock.calls.GetImage = append(mock.calls.GetImage, callInfo)
	mock.lockGetImage.Unlock()
	return mock.GetImageFunc(ctx, userAuthToken, serviceAuthToken, collectionID, imageID)
}

// GetImageCalls gets all the calls that were made to GetImage.
// Check the length with:
//
//	len(mockedImageAPIClient.GetImageCalls())
func (mock *ImageAPIClientMock) GetImageCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ImageID          string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
	}
	mock.lockGetImage.RLock()
	calls = mock.calls.GetImage
	mock.lockGetImage.RUnlock()
	return calls
}

// PostDownloadVariant calls PostDownloadVariantFunc.
func (mock *ImageAPIClientMock) PostDownloadVariant(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.NewImageDownload) (image.ImageDownload, error) {
	if mock.PostDownloadVariantFunc == nil {
		panic("ImageAPIClientMock.PostDownloadVariantFunc: method is nil but ImageAPIClient.PostDownloadVariant was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
		Data             image.NewImageDownload
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ImageID:          imageID,
		Data:             data,
	}
	mock.lockPostDownloadVariant.Lock()
	mock.calls.PostDownloadVariant = append(mock.calls.PostDownloadVariant, callInfo)
	mock.lockPostDownloadVariant.Unlock()
	return mock.PostDownloadVariantFunc(ctx, userAuthToken, serviceAuthToken, collectionID, imageID, data)
}

// PostDownloadVariantCalls gets all the calls that were made to PostDownloadVariant.
// Check the length with:
//
//	len(mockedImageAPIClient.PostDownloadVariantCalls())
func (mock *ImageAPIClientMock) PostDownloadVariantCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ImageID          string
	Data             image.NewImageDownload
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
		Data             image.NewImageDownload
	}
	mock.lockPostDownloadVariant.RLock()
	calls = mock.calls.PostDownloadVariant
	mock.lockPostDownloadVariant.RUnlock()
	return calls
}

// PutDownloadVariant calls PutDownloadVariantFunc.
func (mock *ImageAPIClientMock) PutDownloadVariant(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, variant string, data image.ImageDownload) (image.ImageDownload, error) {
	if mock.PutDownloadVariantFunc == nil {
		panic("ImageAPIClientMock.PutDownloadVariantFunc: method is nil but ImageAPIClient.PutDownloadVariant was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
		Variant          string
		Data             image.ImageDownload
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ImageID:          imageID,
		Variant:          variant,
		Data:             data,
	}
	mock.lockPutDownloadVariant.Lock()
	mock.calls.PutDownloadVariant = append(mock.calls.PutDownloadVariant, callInfo)
	mock.lockPutDownloadVariant.Unlock()
	return mock.PutDownloadVariantFunc(ctx, userAuthToken, serviceAuthToken, collectionID, imageID, variant, data)
}

// PutDownloadVariantCalls gets all the calls that were made to PutDownloadVariant.
// Check the length with:
//
//	len(mockedImageAPIClient.PutDownloadVariantCalls())
func (mock *ImageAPIClientMock) PutDownloadVariantCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ImageID          string
	Variant          string
	Data             image.ImageDownload
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
		Variant          string
		Data             image.ImageDownload
	}
	mock.lockPutDownloadVariant.RLock()
	calls = mock.calls.PutDownloadVariant
	mock.lockPutDownloadVariant.RUnlock()
	return calls
}

// PutImage calls PutImageFunc.
func (mock *ImageAPIClientMock) PutImage(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, imageID string, data image.Image) (image.Image, error) {
	if mock.PutImageFunc == nil {
		panic("ImageAPIClientMock.PutImageFunc: method is nil but ImageAPIClient.PutImage was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
		Data             image.Image
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		CollectionID:     collectionID,
		ImageID:          imageID,
		Data:             data,
	}
	mock.lockPutImage.Lock()
	mock.calls.PutImage = append(mock.calls.PutImage, callInfo)
	mock.lockPutImage.Unlock()
	return mock.PutImageFunc(ctx, userAuthToken, serviceAuthToken, collectionID, imageID, data)
}

// PutImageCalls gets all the calls that were made to PutImage.
// Check the length with:
//
//	len(mockedImageAPIClient.PutImageCalls())
func (mock *ImageAPIClientMock) PutImageCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	CollectionID     string
	ImageID          string
	Data             image.Image
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		CollectionID     string
		ImageID          string
		Data             image.Image
	}
	mock.lockPutImage.RLock()
	calls = mock.calls.PutImage
	mock.lockPutImage.RUnlock()
	return calls
}
