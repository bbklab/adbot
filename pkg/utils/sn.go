package utils

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/bbklab/adbot/pkg/cmd"
)

// SerialNumber format a given string to Serial-Number like text
func SerialNumber(data string) (sn string) {
	hash := sha1.Sum([]byte(data))
	sha1sum := strings.ToUpper(fmt.Sprintf("%x", hash)) // ->  hex & upper case
	for i := 0; i+5 <= len(sha1sum); i += 5 {
		sn += sha1sum[i:i+5] + "-"
	}
	return strings.TrimSuffix(sn, "-")
}

// GetHardwareSerialNumber obtain current system's serial number ...
// - linux:   dmidecode -t system | grep Serial
// - windows: wmic bios get serialnumber
// - OS X:    ioreg -l | grep IOPlatformSerialNumber
func GetHardwareSerialNumber() (string, error) {
	return GetFromDMISystemTable("Serial Number: ")
}

// GetHardwareProductName obtain current system's hardware model & manufacturer ...
func GetHardwareProductName() (string, error) {
	manufacturer, _ := GetFromDMISystemTable("Manufacturer: ")
	productname, err := GetFromDMISystemTable("Product Name: ")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s (%s)", productname, manufacturer), nil
}

// GetFromDMISystemTable is exported
func GetFromDMISystemTable(keyname string) (string, error) {
	stdout, stderr, err := cmd.RunCmd(nil, "dmidecode", "-t", "system")
	if err != nil {
		return "", fmt.Errorf("%v - %s", err, stderr)
	}

	var (
		reader = bufio.NewReader(bytes.NewBuffer([]byte(stdout)))
	)

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, []byte(keyname)) {
			continue
		}

		sn := string(bytes.TrimPrefix(line, []byte(keyname)))
		if sn == "" {
			continue
		}

		return sn, nil
	}

	return "", errors.New("Hardware Serial Number Not Found")
}
