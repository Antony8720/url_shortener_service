package user

import (
	"crypto/aes"
	"encoding/hex"
	"github.com/google/uuid"
)

const key = "aqHC5SN1UQ!mWRrMyJF86lbgo.3Yjein"

type User struct {
	UserID uuid.UUID
}

func New() User {
	return User{UserID: uuid.New()}
}

func (u *User) UserEncrypt() ([]byte, error) {
	aesBlock, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	byteUser, err := u.UserID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encryptedUser := make([]byte, aes.BlockSize)
	aesBlock.Encrypt(encryptedUser, byteUser)
	return encryptedUser, nil
}

func (u *User) UserEncryptEncodeToString() (string, error) {
	eb, err := u.UserEncrypt()
	if err != nil {
		return "", nil
	}
	return hex.EncodeToString(eb), nil
}

func (u *User) UserDecrypt(b []byte) error {
	aesBlock, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}
	decryptedUser := make([]byte, len(uuid.UUID{}))
	aesBlock.Decrypt(decryptedUser, b)

	userID := uuid.New()
	err = userID.UnmarshalBinary(decryptedUser)
	if err != nil {
		return err
	}
	u.UserID = userID
	return nil
}

func (u *User) UserDecryptDecodeFromString(s string) error {
	b, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	return u.UserDecrypt(b)
}
