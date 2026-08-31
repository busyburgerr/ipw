package httpx

import (
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// UploadedImage is a validated image read from a multipart form field.
type UploadedImage struct {
	Data        []byte
	ContentType string
	Ext         string
}

var imageExtByType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// ReadImage extracts an image from the given form field, enforcing a max size
// (bytes) and an allowlist of image content types (sniffed, not trusted from the
// client).
func ReadImage(c *fiber.Ctx, field string, maxBytes int64) (*UploadedImage, error) {
	fh, err := c.FormFile(field)
	if err != nil {
		return nil, ErrBadRequest("missing file field: " + field)
	}
	if fh.Size > maxBytes {
		return nil, ErrBadRequest("file too large")
	}
	f, err := fh.Open()
	if err != nil {
		return nil, ErrBadRequest("cannot read upload")
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return nil, ErrBadRequest("cannot read upload")
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrBadRequest("file too large")
	}

	ct := http.DetectContentType(data)
	ext, ok := imageExtByType[ct]
	if !ok {
		return nil, ErrBadRequest("unsupported image type: " + ct)
	}
	return &UploadedImage{Data: data, ContentType: ct, Ext: ext}, nil
}
