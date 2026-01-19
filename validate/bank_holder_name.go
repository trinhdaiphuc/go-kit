package validate

import (
	"errors"
	"strings"

	"github.com/agnivade/levenshtein"

	"github.com/trinhdaiphuc/go-kit/vi"
)

var (
	ErrTargetNameEmpty        = errors.New("target name is empty")
	ErrInputNameEmpty         = errors.New("input name is empty")
	ErrInvalidInputNameLength = errors.New("invalid input name length")
	ErrInvalidLevenshtein     = errors.New("invalid Levenshtein distance")
	ErrInvalidFirstName       = errors.New("invalid first name")
)

// BankHolderName validates the bank holder name
// inputName: the name that needs to be validated
// targetName: the name that needs to be compared with. Ex: ekyc full name
// levenshteinDistance: the maximum distance between two strings to be considered as equal.
// See https://en.wikipedia.org/wiki/Levenshtein_distance
func BankHolderName(inputName, targetName string, levenshteinDistance int) error {
	inputName = strings.TrimSpace(vi.RemoveAccent(vi.NormalizeText(inputName)))   // Remove accent in full name
	targetName = strings.TrimSpace(vi.RemoveAccent(vi.NormalizeText(targetName))) // Remove accent in full name

	if isEmptyOrOnlyOneWord(inputName) {
		return ErrInputNameEmpty
	}

	if isEmptyOrOnlyOneWord(targetName) {
		return ErrTargetNameEmpty
	}

	// Execute rule follow ticket PCFPN-6844

	// Rule 1: input name length must <= target name length
	if len(inputName) > len(targetName) {
		return ErrInvalidInputNameLength
	}

	// Rule 2: Follow Levenshtein Distance algorithm
	distance := levenshtein.ComputeDistance(strings.ToUpper(inputName), strings.ToUpper(targetName))
	if distance > levenshteinDistance {
		return ErrInvalidLevenshtein
	}

	// Rule 3: Holder name's first name must be equal to target name's first name
	targetFirstName := getFirstName(targetName)
	inputFirstName := getFirstName(inputName)

	if len(targetFirstName) == 0 || len(inputFirstName) == 0 || !strings.EqualFold(targetFirstName, inputFirstName) {
		return ErrInvalidFirstName
	}

	return nil
}

func getFirstName(name string) string {
	nameArr := strings.Split(name, " ")
	if len(nameArr) == 0 {
		return ""
	}
	return nameArr[len(nameArr)-1]
}

func isEmptyOrOnlyOneWord(name string) bool {
	if name == "" || len(strings.Split(name, " ")) == 1 {
		return true
	}
	return false
}
