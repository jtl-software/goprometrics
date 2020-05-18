package main

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/korovkin/limiter"
	"math/rand"
	"net/http"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

var client = &http.Client{}

func main() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	limit := limiter.NewConcurrencyLimiter(5)
	for i := 1; i <= 5000000; i++ {
		limit.Execute(func() {
			h, _ := uuid.NewUUID()
			hash := h.String()
			url := "http://127.0.0.1:9111/ucount/my_namespace/uc/" + hash

			request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader([]byte{}))
			if err != nil {
				panic(err)
			}
			if i%10000 == 0 {
				//print(".")
				println(url)
			}
			count(err, request)
		})

	}
	limit.Wait()
}

func count(err error, request *http.Request) {
	do, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	if do.StatusCode > 300 {
		panic(do.StatusCode)
	}
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		k := seededRand.Intn(len(charset))
		if k < 0 {
			println("k < 0", k)
		}
		b[i] = charset[k]
	}
	return string(b)
}
