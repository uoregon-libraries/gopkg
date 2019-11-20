package pdf

import (
	"fmt"
	"strings"
)

// Image holds a PDF's image data (as parsed from the "pdfimages" command)
type Image struct {
	Page     string
	Num      string
	Type     string
	Width    string
	Height   string
	Color    string
	Comp     string
	BPC      string
	Encoding string
	Interp   string
	Object   string
	ID       string
	XPPI     string
	YPPI     string
	Size     string
	Ratio    string
}

// ImageInfo returns an array of image data by reading the images in the given
// PDF with "pdfimages -list".  The Image data is split into fields for
// convenience, but it is up to the caller to convert to valid int/float
// values, as pdfimages will return non-numeric results in a variety of
// situations.
func ImageInfo(path string) (images []*Image, err error) {
	var lines []string
	lines, err = pdfimagesLines(path)
	if err != nil {
		return nil, err
	}

	for i, line := range lines {
		var image, err = newImageFromRaw(line)
		if err != nil {
			return nil, fmt.Errorf("pdf.Images(): line %d of `pdfimages -list %s`: %s", i+1, path, err)
		}
		images = append(images, image)
	}

	return images, nil
}

func newImageFromRaw(line string) (image *Image, err error) {
	var parts = &str{d: strings.Fields(line)}
	if parts.len() < 15 {
		return nil, fmt.Errorf("too few fields (%q: %#v)", line, parts)
	}

	image = new(Image)

	// Pull the "object" and "id" fields first since the object value can change
	// how many fields we expect
	image.Object = parts.popat(10)
	if image.Object != "[inline]" {
		image.ID = parts.popat(10)
	}

	if parts.len() != 14 {
		return nil, fmt.Errorf("incorrect number of fields (%d after preprocessing: %#v) in line %q", parts.len(), parts, line)
	}

	image.Page = parts.pop()
	image.Num = parts.pop()
	image.Type = parts.pop()
	image.Width = parts.pop()
	image.Height = parts.pop()
	image.Color = parts.pop()
	image.Comp = parts.pop()
	image.BPC = parts.pop()
	image.Encoding = parts.pop()
	image.Interp = parts.pop()
	image.XPPI = parts.pop()
	image.YPPI = parts.pop()
	image.Size = parts.pop()
	image.Ratio = parts.pop()

	return image, err
}
