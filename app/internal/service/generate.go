package service

import "math/rand"

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generate(length int) string {
	res := make([]byte, length)
	for i := range res {
		res[i] = charset[rand.Intn(len(charset))]
	}
	return string(res)
}
