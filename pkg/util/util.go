package util

import (
	"boralabs/config"
	"boralabs/pkg/notification"
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"os"
	"runtime"
	"time"
)

var (
	KstZone *time.Location
)

func init() {
	var err error
	KstZone, err = time.LoadLocation("Asia/Seoul")
	if err != nil {
		panic(err)
	}
}

func NowInKst() time.Time {
	// kst
	return time.Now().In(KstZone)
}

func Log(message string) {
	log.Println(message)
}

func ErrorLog(err error) {
	if err != nil {
		log.Println(err)
		notification.SendAll(err.Error())
	}
}

func IsDebug() bool {
	return os.Getenv("DEBUG") != "" || config.C.GetBool("debug") == true
}

// ReadJSONFile reads a JSON file and decodes it into the given placeholder.
func ReadJSONFile(v any, path string) error {
	fp, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %s / %w", path, err)
	}

	r := bufio.NewReader(fp)
	dec := json.NewDecoder(r)
	if err = dec.Decode(v); err != nil {
		return fmt.Errorf("failed Decode: %w", err)
	}

	return nil
}

func GenerateEventHash(eventSignature string) common.Hash {
	eventSignatureBytes := []byte(eventSignature)
	return crypto.Keccak256Hash(eventSignatureBytes)
}

func ConvArrayToStringArr[T any](data []T) (arr []string) {
	switch dArr := any(data).(type) {
	case []*big.Int:
		for _, d := range dArr {
			arr = append(arr, d.String())
		}
	case []common.Address:
		for _, d := range dArr {
			arr = append(arr, d.String())
		}
	case []common.Hash:
		for _, d := range dArr {
			arr = append(arr, d.String())
		}
	}
	return
}

func PrintStackTrace() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	log.Println("Stack trace:")
	log.Println(string(buf[:n]))
}
