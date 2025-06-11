package headers

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	s := string(data)
	contains := strings.Contains(s, "\r\n")
	if contains == false {
		return 0, false, nil
	}
	if strings.HasPrefix(s, "\r\n") {
		return 2, true, nil
	}
	keyValue := strings.TrimSpace(strings.Split(s, "\r\n")[0])
	parts := strings.SplitN(keyValue, ":", 2)
	key := parts[0]
	if unicode.IsSpace(rune(key[len(key)-1])) {
		return 0, false, errors.New("Incorrect fieldname format")
	}
	key = strings.TrimSpace(parts[0])
	if regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+-.^_|~`]*$").MatchString(key) == false {
		return 0, false, errors.New("Field name must only contain valid characters")
	}
	key = strings.ToLower(key)
	value := strings.TrimSpace(parts[1])
	_, exists := h[key]
	if exists {
		h[key] = h[key] + ", " + value
	} else {
		h[key] = value
	}
	return len(strings.Split(s, "\r\n")[0]) + 2, false, nil
}
