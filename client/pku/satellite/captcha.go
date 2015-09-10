package pku

// #cgo CXXFLAGS: -O3
// #include "nzkcaptcha.h"
import "C"
import (
	"image"
	"unsafe"
)

func Identify(im image.Image) string {
	rect := im.Bounds()
	baseW, baseH := rect.Min.X, rect.Min.Y
	w, h := rect.Max.X-baseW, rect.Max.Y-baseH

	imbit := make([]uint32, w)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c, _, _, _ := im.At(x+baseW, y+baseH).RGBA()
			if c > 32767 {
				imbit[x] |= 1 << uint(y)
			}
		}
	}

	res := make([]byte, 5)
	ptr := (*C.char)(unsafe.Pointer(&res[0]))
	C.identify(C.int(h), C.int(w), (*C.int)(unsafe.Pointer(&imbit[0])), ptr)
	return C.GoString(ptr)
}
