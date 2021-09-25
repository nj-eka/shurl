package hashid_tokenizer

import (
	"fmt"
	"github.com/nj-eka/shurl/app"
	"github.com/nj-eka/shurl/config"
	"log"
	"testing"
)

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

var tokenizer app.Tokenizer
var id2key = map[int]string{
	0:          "EZead",
	1:          "EdGed",
	2:          "XBa80",
	1000:       "ZeEMB",
	1000000:    "ejM7A",
	1000000000: "e1lLANM",
	MaxInt:     "DApEj4wbneowA",
}

//salt: "ecafbaf0-1bcc-11ec-9621-0242ac130002"
//min-length: 5
//alphabet: "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
func init() {
	var err error
	tokenizer, err = NewHashidTokenizer(&config.HashidTokenizerConfig{
		Salt:      "ecafbaf0-1bcc-11ec-9621-0242ac130002",
		MinLength: 5,
		Alphabet:  "0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func TestHashidTokenizer_GetKey(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"encode -1", args{-1}, "", true},
		{"encode0", args{0}, id2key[0], false},
		{"encode1", args{1}, id2key[1], false},
		{"encode2", args{2}, id2key[2], false},
		{"encode1000", args{1000}, id2key[1000], false},
		{"encode1000000", args{1000000}, id2key[1000000], false},
		{"encode1000000000", args{1000000000}, id2key[1000000000], false},
		{"encodeMaxInt", args{MaxInt}, id2key[MaxInt], false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.Encode(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashidTokenizer_GetId(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"decode -> fail", args{}, -1, true},
		{"decode -> 0", args{id2key[0]}, 0, false},
		{"decode -> 1", args{id2key[1]}, 1, false},
		{"decode -> 2", args{id2key[2]}, 2, false},
		{"decode -> 1000", args{id2key[1000]}, 1000, false},
		{"decode -> 1000000", args{id2key[1000000]}, 1000000, false},
		{"decode -> 1000000000", args{id2key[1000000000]}, 1000000000, false},
		{"decode MaxInt", args{id2key[MaxInt]}, MaxInt, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizer.Decode(tt.args.key)
			fmt.Println(got, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
