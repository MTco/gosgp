package main

import "hash"

const (
	WASH_ROUNDS         = 10
	MIN_PASSWORD_LENGTH = 4
)

type SGP interface {
	Hasher() hash.Hash
	MaxLength() int
	ZeroBytes()
	PwBuf() []byte   // used by generatePass
	HashBuf() []byte // used by Generate()
	FixPadding([]byte)
}

func SupergenPass(out []byte, hasher SGP, password, domain []byte) error {
	return generatePass(out, hasher, password, []byte(":"), domain)
}

func generatePass(out []byte, sgp SGP, pw_parts ...[]byte) error {

	if len(out) > sgp.MaxLength() {
		return errorRequestTooLong(len(out), sgp.MaxLength())
	}

	pw := sgp.PwBuf()
	digest := sgp.HashBuf()
	defer zeroBytes(digest)

	hashSlices(digest, sgp.Hasher(), pw_parts...)
	sgpBase64(pw, digest, _SGP_BASE64_ALPHABET)
	sgp.FixPadding(pw)

	for round := 1; round < WASH_ROUNDS; round += 1 {
		hashSlices(digest, sgp.Hasher(), pw)
		sgpBase64(pw, digest, _SGP_BASE64_ALPHABET)
		sgp.FixPadding(pw)
	}

	for !passwordIsValid(pw[:len(out)]) {
		hashSlices(digest, sgp.Hasher(), pw)
		sgpBase64(pw, digest, _SGP_BASE64_ALPHABET)
		sgp.FixPadding(pw)
	}

	copy(out, pw)
	return nil
}

// returns true only if:
//
// 1. 'password' must start with a lowercase letter [a-z].
// 2. 'password' must contain at least one uppercase letter [A-Z].
// 3. 'password' must contain at least one numeral [0-9].
//
// see 'var validatePassword = function (str, length) { ... }' in
// github.com/chriszarate/supergenpass-lib/blob/master/supergenpass-lib.js
func passwordIsValid(password []byte) bool {

	if len(password) == 0 {
		return false
	}

	if !(password[0] >= 'a' && password[0] <= 'z') {
		return false
	}

	var has_digit, has_LETTER bool

	for i := 0; !(has_digit && has_LETTER) && i < len(password); i++ {
		c := password[i]
		if c >= '0' && c <= '9' {
			has_digit = true
		} else if c >= 'A' && c <= 'Z' {
			has_LETTER = true
		}
	}

	return (has_digit && has_LETTER)
}

func hashSlices(out []byte, hasher hash.Hash, slices ...[]byte) {
	defer hasher.Reset()
	hasher.Reset()
	for i := range slices {
		hasher.Write(slices[i])
	}
	hasher.Sum(out)
}
