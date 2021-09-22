package common

import (
    "math/rand"
    "time"
)

func RandInt(min, max int) int {
 
    s := rand.NewSource(time.Now().UnixNano())
    r := rand.New(s)
	return r.Intn(max - min) + min

}