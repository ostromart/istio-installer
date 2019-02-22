package fetch

import (
	"context"
	"io"
	"net/http"
	"os"
"cloud.google.com/go/storage"
)


func GetFileHTTP(srcURL string, dstFilepath string) error {
	out, err := os.Create(dstFilepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(srcURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetFileGCS(ctx context.Context, bucketName, srcFilepath, dstFilepath string) error {
	out, err := os.Create(dstFilepath)
	if err != nil {
		return err
	}
	defer out.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)

	in, err := bucket.Object(srcFilepath).NewReader(ctx)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}
