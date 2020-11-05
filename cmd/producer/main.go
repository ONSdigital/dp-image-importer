package main

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/image"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/log.go/log"
	"os"
)

const (
	s3UploadPath = "578-battenburgpng"
	collectionId = "123"
)

func main() {

	ctx := context.Background()
	log.Namespace = "dp-dataset-exporter"

	cfg, err := config.Get()
	if err != nil {
		log.Event(ctx, "error getting cfg", log.FATAL, log.Error(err))
		os.Exit(1)
	}
	log.Event(ctx, "loaded cfg", log.INFO, log.Data{"cfg": cfg})

	// Create image via api
	imageApi := image.NewAPIClient(cfg.ImageAPIURL)

	newImage := image.NewImage{
		CollectionId: collectionId,
		State:        "created",
		Filename:     "tiny.png",
		Type:         "pixel",
	}
	image, err := imageApi.PostImage(ctx, "", cfg.ServiceAuthToken, collectionId, newImage)
	if err != nil {
		log.Event(ctx, "fatal error trying to create image record in API", log.FATAL, log.Error(err), log.Data{"new_image": newImage})
		os.Exit(1)
	}
	image.State = "uploaded"
	image.Upload.Path = s3UploadPath
	image, err = imageApi.PutImage(ctx, "", cfg.ServiceAuthToken, collectionId, image.Id, image)
	if err != nil {
		log.Event(ctx, "fatal error trying to update image record with upload info", log.FATAL, log.Error(err), log.Data{"new_image": newImage})
		os.Exit(1)
	}
}
