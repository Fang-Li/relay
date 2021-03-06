package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
)

type Cache struct {
	History []Cmd
}

// encrypt cache just keep direct tamper away, anyone run relay can still view the history
var key = []byte("lootilcloocyrevasiyaleremeveileb")

func loadCache() (c Cache, err error) {
	data, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		return c, err
	}
	if data, err = decrypt(key, data); err != nil {
		return
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return
	}
	return
}

func (cache *Cache) AppendHistory(c Cmd) {
	index := -1
	length := len(cache.History)
	for i, cc := range cache.History {
		if cc.Equals(c) {
			index = i
			break
		}
	}
	if index >= 0 {
		for i := index; i < length-1; i++ {
			cache.History[i] = cache.History[i+1]
		}
		if index != length-1 {
			cache.History[length-1] = c
		}
	} else if index == -1 {
		cache.History = append(cache.History, c)
	}
}

func saveCache(c Cache) {
	// keep 200 history
	hmax := 200
	if l := len(c.History); l > hmax {
		c.History = c.History[(l - hmax):l]
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(c)
	data, err := encrypt(key, buffer.Bytes())
	if err != nil {
		return
	}
	ioutil.WriteFile(cacheFile, data, 0644)
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}
