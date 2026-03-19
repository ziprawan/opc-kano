package pddikti

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kano/internal/config"
	"net/http"
	"net/url"
	"strings"
)

var ErrNoKeyOrIv = errors.New("no specified key or iv")

var BASE_URL string = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 97, 112, 105, 45, 112, 100, 100, 105, 107, 116, 105, 46, 107, 101, 109, 100, 105, 107, 116, 105, 115, 97, 105, 110, 116, 101, 107, 46, 103, 111, 46, 105, 100, 47, 112, 101, 110, 99, 97, 114, 105, 97, 110, 47, 101, 110, 99, 47, 97, 108, 108, 47})

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("given data is empty")
	}
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("given data has invalid block size")
	}

	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize {
		return nil, fmt.Errorf("invalid padding")
	}

	for _, v := range data[len(data)-padLen:] {
		if int(v) != padLen {
			return nil, fmt.Errorf("invalid padding content")
		}
	}

	return data[:len(data)-padLen], nil
}

func decryptSearchResult(body, key, iv []byte) ([]byte, error) {
	if len(key) == 0 || len(iv) == 0 {
		return nil, ErrNoKeyOrIv
	}
	cipherText, err := base64.StdEncoding.DecodeString(string(body[1 : len(body)-2]))
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid cipherText block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(cipherText))
	mode.CryptBlocks(decrypted, cipherText)
	decrypted, err = pkcs7Unpad(decrypted, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func buildUrl(path ...string) string {
	escaped := make([]string, len(path))
	for i := range path {
		escaped[i] = url.PathEscape(path[i])
	}

	return BASE_URL + strings.Join(escaped, "/")
}

func fetch(url string) (*http.Response, error) {
	origUrl := BASE_URL[:8] + BASE_URL[12:41]

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Origin", origUrl)
	req.Header.Set("Referer", origUrl)

	return client.Do(req)
}

func Search(query string) (*DiddySearchResult, error) {
	key := config.GetConfig().PddiktiKey
	iv := config.GetConfig().PddiktiIv

	url := BASE_URL + url.PathEscape(query)
	resp, err := fetch(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("expected HTTP code 200, got %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decrypted, err := decryptSearchResult(body, key, iv)
	if err != nil {
		return nil, err
	}

	var res DiddySearchResult
	err = json.Unmarshal(decrypted, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
