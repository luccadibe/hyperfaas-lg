package internal

import (
	"encoding/base64"
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

type DataProvider interface {
	GetData() []byte
}

type EchoDataProvider struct {
	minSize int
	maxSize int
	rand    *rand.Rand
}

func NewEchoDataProvider(minSize int, maxSize int) *EchoDataProvider {
	return &EchoDataProvider{
		minSize: minSize,
		maxSize: maxSize,
		rand:    rand.New(rand.NewSource(123)),
	}
}

func (e *EchoDataProvider) GetData() []byte {
	minSize := e.minSize
	if minSize < 8 {
		minSize = 8
	}
	maxSize := e.maxSize
	if maxSize < minSize {
		maxSize = minSize
	}
	size := e.rand.Intn((maxSize-minSize)+1) + minSize
	size = size &^ 7 // round down to nearest multiple of 8
	if size < 8 {
		size = 8
	}
	data := make([]byte, size)
	for i := 0; i < size; i += 8 {
		u := e.rand.Uint64()
		data[i+0] = byte(u)
		data[i+1] = byte(u >> 8)
		data[i+2] = byte(u >> 16)
		data[i+3] = byte(u >> 24)
		data[i+4] = byte(u >> 32)
		data[i+5] = byte(u >> 40)
		data[i+6] = byte(u >> 48)
		data[i+7] = byte(u >> 56)
	}
	return data
}

/*
this is what bfs json expects

	type InputData struct {
		Size int
	}
*/
type BFSJSONDataProvider struct {
	minSize int
	maxSize int
	rand    *rand.Rand
}

func NewBFSJSONDataProvider(minSize int, maxSize int) *BFSJSONDataProvider {
	return &BFSJSONDataProvider{
		minSize: minSize,
		maxSize: maxSize,
		rand:    rand.New(rand.NewSource(123)),
	}
}

func (b *BFSJSONDataProvider) GetData() []byte {
	size := b.rand.Intn((b.maxSize-b.minSize)+1) + b.minSize
	buf := make([]byte, 0, 16) // 16 is enough for {"Size":<int>}
	buf = append(buf, '{', '"', 'S', 'i', 'z', 'e', '"', ':')
	buf = strconv.AppendInt(buf, int64(size), 10)
	buf = append(buf, '}')
	return buf
}

func fetchThumbnailerImage() ([]byte, error) {
	resp, err := http.Get("http://picsum.photos/1920/1080")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

/*
this is what thumbnailer json expects

	type InputData struct {
		Image  []byte `json:"image"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
*/
type ThumbnailerJSONDataProvider struct {
	rand  *rand.Rand
	image []byte
}

func NewThumbnailerJSONDataProvider() *ThumbnailerJSONDataProvider {
	img, err := fetchThumbnailerImage()
	if err != nil {
		// If image fetch fails, panic or handle as needed. Here, panic for simplicity.
		panic(err)
	}
	return &ThumbnailerJSONDataProvider{
		rand:  rand.New(rand.NewSource(123)),
		image: img,
	}
}

func (t *ThumbnailerJSONDataProvider) GetData() []byte {
	width := t.rand.Intn(1440) + 1
	height := t.rand.Intn(900) + 1

	b64Len := base64.StdEncoding.EncodedLen(len(t.image))
	buf := make([]byte, 0, b64Len+64)
	buf = append(buf, '{', '"', 'i', 'm', 'a', 'g', 'e', '"', ':', '"')
	b64 := make([]byte, b64Len)
	base64.StdEncoding.Encode(b64, t.image)
	buf = append(buf, b64...)
	buf = append(buf, '"', ',')
	buf = append(buf, '"', 'w', 'i', 'd', 't', 'h', '"', ':')
	buf = strconv.AppendInt(buf, int64(width), 10)
	buf = append(buf, ',')
	buf = append(buf, '"', 'h', 'e', 'i', 'g', 'h', 't', '"', ':')
	buf = strconv.AppendInt(buf, int64(height), 10)
	buf = append(buf, '}')
	return buf
}
