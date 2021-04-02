package main

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"cloud.google.com/go/datastore"
	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
)

// helper funcs

func deleteElement(sli []Bot, del Bot) []Bot {
	var rSli []Bot
	for _, e := range sli {
		if e.K.ID != del.K.ID {
			rSli = append(rSli, e)
		}
	}
	return rSli
}

func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateEncryptKey(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

var iv = []byte{34, 12, 55, 11, 10, 39, 16, 47, 87, 53, 88, 98, 66, 40, 14, 05}

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func encrypt(key, text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
	// block, err := aes.NewCipher([]byte(key))
	// if err != nil {
	// 	panic(err)
	// }
	// plaintext := []byte(text)
	// cfb := cipher.NewCFBEncrypter(block, iv)
	// ciphertext := make([]byte, len(plaintext))
	// cfb.XORKeyStream(ciphertext, plaintext)
	// return encodeBase64(ciphertext)
}

func decrypt(key, text string) string {
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		panic(err)
	}
	return string(data)
	// block, err := aes.NewCipher([]byte(key))
	// if err != nil {
	// 	panic(err)
	// }
	// ciphertext := decodeBase64(text)
	// cfb := cipher.NewCFBEncrypter(block, iv)
	// plaintext := make([]byte, len(ciphertext))
	// cfb.XORKeyStream(plaintext, ciphertext)
	// return string(plaintext)
}

var nums = []rune("1234567890")

func generateWebhookID(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = nums[rand.Intn(len(nums))]
	}
	return string(b)
}

func initRedis() {
	if redisHost == "" {
		redisHost = "127.0.0.1"
		fmt.Println("Env var nil, using redis dev address -- " + redisHost)
	}
	if redisPort == "" {
		redisPort = "6379"
		fmt.Println("Env var nil, using redis dev port -- " + redisPort)
	}
	fmt.Println("Connected to Redis on " + redisHost + ":" + redisPort)
	rdb = redis.NewClient(&redis.Options{
		Addr: redisHost + ":" + redisPort,
	})
}

func setupCORS(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") //temp
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, auth")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func authenticateUser(req loginReq) (bool, User) {
	// get user with id/email
	var userWithEmail User
	var query *datastore.Query
	if req.Email != "" {
		query = datastore.NewQuery("User").
			Filter("Email =", req.Email)
	} else if req.ID != "" {
		i, _ := strconv.Atoi(req.ID)
		key := datastore.IDKey("User", int64(i), nil)
		query = datastore.NewQuery("User").
			Filter("__key__ =", key)
	} else {
		return false, User{}
	}

	t := client.Run(ctx, query)
	_, error := t.Next(&userWithEmail)
	if error != nil {
		// Handle error.
	}
	// check password hash and return
	return CheckPasswordHash(req.Password, userWithEmail.Password), userWithEmail
}
