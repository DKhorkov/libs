package security

import "encoding/base64"

// Encode encodes data for security purpose. For example, to send via HTTP.
func Encode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Decode decodes encoded data, to get it's original value.
func Decode(encoded string) ([]byte, error) {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}
