package processing

import (
	"image"
	"math"
	"sync"

	"backend-election/internal/pkg/fingerprinting/helpers"
	"backend-election/internal/pkg/fingerprinting/matrix"
	"backend-election/internal/pkg/fingerprinting/types"
)

func Normalize(in, out *matrix.M, meta types.Metadata) {
	helpers.RunInParallel(in, 0, func(wg *sync.WaitGroup, bounds image.Rectangle) {
		doNormalize(in, out, bounds, meta.MinValue, meta.MaxValue)
		wg.Done()
	})
}
func doNormalize(in, out *matrix.M, bounds image.Rectangle, min, max float64) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := in.At(x, y)
			normalizedPixel := math.MaxUint8 * (pixel - min) / (max - min)
			out.Set(x, y, normalizedPixel)
		}
	}
}
