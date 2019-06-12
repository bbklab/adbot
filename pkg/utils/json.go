package utils

import (
	"encoding/json"
	"io"
	"os"
)

// PrettyJSON is exported
func PrettyJSON(w io.Writer, data interface{}) error {
	if w == nil {
		w = io.Writer(os.Stdout)
	}

	bs, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	w.Write(append(bs, '\r', '\n'))
	return nil
}
