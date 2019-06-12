package qrcode

import (
	// dec "github.com/PeterCxy/gozbar" // required: yum -y install zbar zbar-devel,  require CGO building
	enc "github.com/skip2/go-qrcode"
)

// Encode is exported
func Encode(data string) ([]byte, error) {
	return enc.Encode(data, enc.Medium, 256)
}

/*
// Decode is exported
//
// disable this function because it required too much changing on current building steps
//  - enable CGO
//  - building env must installed:  gcc zbar zbar-devel
//  - the result binary dependency lots of .so files
//
func Decode(png []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewBuffer(png))
	if err != nil {
		return "", err
	}

	s := dec.NewScanner()
	s.SetConfig(dec.QRCODE, dec.CFG_ENABLE, 1)

	image := dec.FromImage(img)
	result := s.Scan(image)
	if result < 0 {
		return "", errors.New("error occurred when scanning")
	}

	if result == 0 {
		return "", errors.New("no symbols found")
	}

	var res []string
	image.First().Each(func(item string) {
		res = append(res, item)
	})

	if len(res) == 0 {
		return "", errors.New("no symbols found")
	}

	return res[0], nil
}
*/
