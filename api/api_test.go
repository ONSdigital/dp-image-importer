package api_test

import (
	"context"
	"github.com/gorilla/mux"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-image-importer/api"
	"github.com/ONSdigital/dp-image-importer/api/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		api := api.Setup(ctx, r, &mock.VaultClienterMock{}, &mock.S3ClienterMock{}, &mock.S3ClienterMock{}, &mock.ImageAPIClienterMock{})

		Convey("When created the following routes should have been added", func() {
			// Replace the check below with any newly added api endpoints
			So(hasRoute(api.Router, "/hello", "GET"), ShouldBeTrue)
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		api := api.Setup(ctx, r, &mock.VaultClienterMock{}, &mock.S3ClienterMock{}, &mock.S3ClienterMock{}, &mock.ImageAPIClienterMock{})

		Convey("When the api is closed any dependencies are closed also", func() {
			err := api.Close(ctx)
			So(err, ShouldBeNil)
			// Check that dependencies are closed here
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, nil)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
