package pvoc

import(
  "math"
)

var gOmegaPiImag []float64 = make([]float64, 31, 31)
var gOmegaPiReal []float64 = make([]float64, 31, 31)

func init() {
  var N uint32 = 2

  for i := 0; i < 31; i++ {
    NFloat := float64(N)
    gOmegaPiImag[i] = math.Sin(twoPi / NFloat)
    gOmegaPiReal[i] = -2 * math.Sin(pi / NFloat) * math.Sin(pi / NFloat)

    N <<= 1
  }
}

// rearranges data array in bit-reversal order in-place
// NOTE: the data array contains data in pairs, so the
// bitreversal operates on data pairs (element i and i + 1)
// and is not exactly a straight bitreversal, if our array is:
// [0, 1, 2, 3, 4, 5, 6, 7] <- array data
//  ____  ____  ____  ____
//   0     1     2     3    <- indexes to bit reverse
func bitReverse(data []float64) {
  var m int

  for i, j := 0, 0; i < len(data); i, j = i + 2, j + m {
    if j > i {
      realTemp := data[j]
      imagTemp := data[j + 1]
      data[j] = data[i]
      data[j + 1] = data[i + 1]
      data[i] = realTemp
      data[i + 1] = imagTemp
    }

    for m = len(data) / 2; m >= 2 && j >= m; m /= 2 {
      j -= m
    }
  }
}

func FFT(data []float64, direction int) {
  bitReverse(data)

  numberData := len(data)
  halfPoints := numberData / 2

  var twoMMax int
  n := 0
  for mMax := 2; mMax < numberData; mMax = twoMMax {
    twoMMax = mMax * 2
    omegaPiReal := gOmegaPiReal[n]

    var omegaPiImag float64

    if direction == Time2Freq {
      omegaPiImag = gOmegaPiImag[n]
    } else {
      omegaPiImag = -gOmegaPiImag[n]
    }
    n++

    omegaReal := 1.0
    omegaImag := 0.0

    for m := 0; m < mMax; m += 2 {
      var imagTemp, realTemp float64
      for i := m; i < numberData; i += twoMMax {
        j := i + mMax
        realTemp = omegaReal * data[j] - omegaImag * data[j + 1]
        imagTemp = omegaReal * data[j + 1] + omegaImag * data[j]
        data[j] = data[i] - realTemp
        data[j + 1] = data[i + 1] - imagTemp
        data[i] += realTemp
        data[i + 1] += imagTemp
      }
      realTemp = omegaReal
      omegaReal = omegaReal * omegaPiReal - omegaImag * omegaPiImag + omegaReal
      omegaImag = omegaImag * omegaPiReal + realTemp * omegaPiImag + omegaImag
    }
  }

  if (direction == Freq2Time) {
    scale := 1.0 / float64(halfPoints);

    for i := 0; i < numberData; i++ {
      data[i] *= scale;
    }
  }
}

// Comment from SoundHack:
// RealFFT - performs fft with only real values and positive frequencies
func RealFFT(data []float64, direction int) {
  points := len(data)
  halfPoints := points / 2

  twoPiOmmax := pi / float64(halfPoints)
  omegaReal := 1.0
  omegaImag := 0.0
  c1 := 0.5

  var c2, xr, xi float64

  if direction == Time2Freq {
    c2 = -0.5
    FFT(data, direction)
    xr = data[0]
    xi = data[1]
  } else {
    c2 = 0.5
    twoPiOmmax = -twoPiOmmax
    xr = data[1]
    xi = 0.0
    data[1] = 0.0
  }

  temp := math.Sin(0.5 * twoPiOmmax)
  omegaPiReal := -2.0 * temp * temp
  omegaPiImag := math.Sin(twoPiOmmax)
  N2p1 := points + 1;

  for i := 0; i <= halfPoints / 2; i++ {
    i1 := i * 2
    i2 := i1 + 1
    i3 := N2p1 - i2
    i4 := i3 + 1

    if i == 0 {
      h1r :=  c1 * (data[i1] + xr)
      h1i :=  c1 * (data[i2] - xi)
      h2r := -c2 * (data[i2] + xi)
      h2i :=  c2 * (data[i1] - xr)
      data[i1] = h1r + omegaReal * h2r - omegaImag * h2i
      data[i2] = h1i + omegaReal * h2i + omegaImag * h2r
      xr =  h1r - omegaReal * h2r + omegaImag * h2i
      xi = -h1i + omegaReal * h2i + omegaImag * h2r
    } else {
      h1r :=  c1 * (data[i1] + data[i3])
      h1i :=  c1 * (data[i2] - data[i4])
      h2r := -c2 * (data[i2] + data[i4])
      h2i :=  c2 * (data[i1] - data[i3])
      data[i1] =  h1r + omegaReal * h2r - omegaImag * h2i
      data[i2] =  h1i + omegaReal * h2i + omegaImag * h2r
      data[i3] =  h1r - omegaReal * h2r + omegaImag * h2i
      data[i4] = -h1i + omegaReal * h2i + omegaImag * h2r
    }
    temp = omegaReal
    omegaReal = omegaReal * omegaPiReal - omegaImag * omegaPiImag + omegaReal
    omegaImag = omegaImag * omegaPiReal + temp * omegaPiImag + omegaImag
  }

  if direction == Time2Freq {
    data[1] = xr
  } else {
    FFT(data, direction)
  }
}
