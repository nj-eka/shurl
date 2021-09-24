package hashid_tokenizer

import (
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	"log"
	"testing"
)

var tz app.Tokenizer

func init(){
	var err error
	tz, err = NewHashidTokenizer(&config.HashidTokenizerConfig{
		Salt:      "ecafbaf0-1bcc-11ec-9621-0242ac130002",
		MinLength: 5,
		Alphabet:  "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	})
	if err != nil{
		log.Fatalln(err)
	}
}

func BenchmarkHashidTokenizer_GetKey(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tz.Encode(i)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkHashidTokenizer_Gets(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key, _ := tz.Encode(i)
		j, err := tz.Decode(key)
		if err != nil {
			b.Error(err)
		}
		if j != i{
			b.Error("fail")
		}
	}
}

//goos: linux
//goarch: amd64
//pkg: github.com/nj-eka/shurl/app/hashid_tokenizer
//cpu: Intel(R) Core(TM) i7-4710HQ CPU @ 2.50GHz
//BenchmarkHashidTokenizer_GetKey
//BenchmarkHashidTokenizer_GetKey-4         836636              1364 ns/op
//BenchmarkHashidTokenizer_Gets
//BenchmarkHashidTokenizer_Gets-4           278976              4309 ns/op
