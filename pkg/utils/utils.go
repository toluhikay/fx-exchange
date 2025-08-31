package utils

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"maps"

	"github.com/go-playground/validator"

	customErrors "github.com/toluhikay/fx-exchange/internal/errors"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GenerateOtp() string {
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	return otp
}

func GeneratePassword() string {
	otp := fmt.Sprintf("%x", rand.Intn(999999999))
	return otp
}

type JSONResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func WriteJson(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	outData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// check the lenth of the header
	if len(headers) > 0 {
		maps.Copy(w.Header(), headers[0])
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(outData)

	return nil
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	// first set max data size
	maxDataBytes := 1024 * 1024 //1mb is my prefered max

	// read the size of the request body to make sure it does not acept more than the specified size
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxDataBytes))

	// create a decoder
	dec := json.NewDecoder(r.Body)

	// make sure decoder does not accept unknown fields
	dec.DisallowUnknownFields()

	// make sure the decoder does not accept if all fields are

	// decode the data
	err := dec.Decode(data)
	if err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("invalid JSON syntax at byte offset %d", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			return fmt.Errorf("invalid type for field %q: expected %s, got %s", unmarshalTypeErr.Field, unmarshalTypeErr.Type, unmarshalTypeErr.Value)
		case errors.As(err, &invalidUnmarshalErr):
			return fmt.Errorf("invalid unmarshal: %v", invalidUnmarshalErr)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			// Extract the unknown field name from the error message
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unrecognized field %s in JSON payload", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("request body exceeds maximum size of %d bytes", maxDataBytes)
		default:
			return fmt.Errorf("failed to decode JSON: %v", err)
		}
	}

	// make sure request contain not more than one json payload
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain one json file")
	}

	// Validate required fields using the validator library
	validate := validator.New()
	err = validate.Struct(data)
	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Collect specific validation errors
			var errorMessages []string
			for _, fieldErr := range validationErrors {
				switch fieldErr.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("missing required field %q", fieldErr.Field()))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("validation failed for field %q: %s", fieldErr.Field(), fieldErr.Tag()))
				}
			}
			return fmt.Errorf("validation errors: %s", strings.Join(errorMessages, "; "))
		}
		return fmt.Errorf("validation failed: %v", err)
	}

	return nil
}

func ErrorJSON(w http.ResponseWriter, err error, status ...int) error {
	// set a default status code in case of  none returned
	statusHeader := http.StatusBadRequest

	// check if the lenght of status is more than 0
	if len(status) > 0 {
		statusHeader = status[0]
	}

	// build the return data
	var payload = JSONResponse{
		Error:   true,
		Message: err.Error(),
	}

	return WriteJson(w, statusHeader, payload)

}

// create a function to encrypt mail that  will be sent to the user's mail which the frontend will send back
// and be decrypted
var key = []byte("32-byte-long-secret-key-12345678") // 32 bytes exactly

func EncryptNewOtp(otpMail string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("error creating block at encrypt", err)
		return "", customErrors.ErrInternalServer
	}

	ciphertext := make([]byte, aes.BlockSize+len(otpMail))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptoRand.Reader, iv); err != nil {
		fmt.Println("error reading at encry", err)
		return "", customErrors.ErrInternalServer
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(otpMail))

	return base64.URLEncoding.EncodeToString(ciphertext), nil

}

// Decrypt decrypts data using AES
func Decrypt(encrypted string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		fmt.Println("error here")
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("err decrypt - 0", err)
		return "", customErrors.ErrInvalidOtp
	}

	if len(ciphertext) < aes.BlockSize {
		return "", customErrors.ErrInvalidOtp
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
