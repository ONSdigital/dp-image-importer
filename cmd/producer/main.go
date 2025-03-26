package main

import (
	"context"
	"os"

	"github.com/ONSdigital/dp-api-clients-go/v2/image"
	"github.com/ONSdigital/dp-image-importer/config"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	s3UploadPath = "578-battenburgpng"
	collectionId = "123"
)

func main() {

	ctx := context.Background()
	log.Namespace = "dp-image-importer-producer"

	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "error getting cfg", err)
		os.Exit(1)
	}
	log.Info(ctx, "loaded cfg", log.Data{"cfg": cfg})

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
		log.Fatal(ctx, "fatal error trying to create image record in API", err, log.Data{"new_image": newImage})
		os.Exit(1)
	}
	image.State = "uploaded"
	image.Upload.Path = s3UploadPath
	image, err = imageApi.PutImage(ctx, "", cfg.ServiceAuthToken, collectionId, image.Id, image)
	if err != nil {
		log.Fatal(ctx, "fatal error trying to update image record with upload info", err, log.Data{"new_image": newImage})
		os.Exit(1)
	}
}
