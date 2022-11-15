package utils

import (
	"errors"
	"unicode"

	"github.com/badoux/checkmail"
	"golang.org/x/crypto/bcrypt"
)

const minEntropy = 42

var (
	ErrorRecordNotFound     = errors.New("record not found")
	ErrorEmptyLogin         = errors.New("you must have email for login")
	ErrorEmptyCredentials   = errors.New("you must have name and surname")
	ErrorEmptyPass          = errors.New("you must have password for you account")
	ErrorUserExists         = errors.New("user already exists")
	ErrorUserNotFound       = errors.New("user not found")
	ErrorUnauthenticated    = errors.New("login or password is uncorrect")
	ErrorOldPassInvalid     = errors.New("old pass isn`t valid")
	ErrorInvalidFormat      = errors.New("invalid format of input string")
	ErrorInvalidEmailFormat = errors.New("invalid email format")
	ErrorPasswordNotMatched = errors.New("bad old password")
	ErrorNewOldPassMatched  = errors.New("new and old password musn`t be matched")
	ErrorBadPassword        = errors.New("password must have 8 chars, at least one uppercase, one lowercase letter, one number and one special char")
	ErrorNoNewData          = errors.New("no data for update")
)

func CheckEmail(input string) bool {
	if err := checkmail.ValidateFormat(input); err != nil {
		return false
	}
	return true
}

// Password validation with OWASP recomendations
// upp: at least one upper case letter.
// low: at least one lower case letter.
// num: at least one digit.
// sym: at least one special character.
// tot: at least eight characters long.
// No empty string or whitespace.
func ValidatePassword(pass string) bool {
	var (
		upp, low, num, sym bool
		tot                uint8
	)

	for _, char := range pass {
		switch {
		case unicode.IsUpper(char):
			upp = true
			tot++
		case unicode.IsLower(char):
			low = true
			tot++
		case unicode.IsNumber(char):
			num = true
			tot++
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			sym = true
			tot++
		default:
			return false
		}
	}

	if !upp || !low || !num || !sym || tot < 8 {
		return false
	}

	return true
}

func EncodingPassword(pass string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(bytes), err
}

func ComparePasswords(pass string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}
