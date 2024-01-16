package docx

import "encoding/json"

// InlineImage represents an InlineImage field. You can then create a placeholder as usual, and give it a
// value with InlineImage as the type, and an image will be rendered in place of the placeholder.
// Data with field type InlineImage will be encoded as InlineImage JSON, specialized to be converted into a DOCX inline
// image by siasn-docx. It is encoded into a string, but contains a valid JSON to build an InlineImage, like
// an image_descriptor field.
//
// ImageDescriptor field must not be empty for this to be a valid InlineImage.
type InlineImage struct {
	// Path to the image that will be loaded by the siasn-docx script.
	ImageDescriptor string
	// Must be a positive value, negative or 0 will be ignored.
	// If unspecified (0 or negative), the width of the image will be the same as the original image.
	Width float32
	// Must be a positive value, negative or 0 will be ignored.
	// If unspecified (0 or negative), the height of the image will be the same as the original image.
	Height float32
}

func (i InlineImage) MarshalJSON() ([]byte, error) {
	raw := map[string]interface{}{
		"type":             "InlineImage",
		"image_descriptor": i.ImageDescriptor,
		"width":            i.Width,
		"height":           i.Height,
	}

	payload, _ := json.Marshal(raw)
	return json.Marshal(string(payload))
}
