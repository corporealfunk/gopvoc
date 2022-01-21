package pvoc

import(
  "math"
  "strings"
)

const twoPi float64 = math.Pi * 2
const pi float64 = math.Pi

type windowFunction func (int) []float64

var WindowFunctions = map[string]windowFunction {
  "hamming": HammingWindow,
  "vonhann": VonHannWindow,
  "kaiser": KaiserWindow,
  "sinc": SincWindow,
  "triangle": TriangleWindow,
  "ramp": RampWindow,
  "rectangle": RectangleWindow,
}

func WindowNames() []string {
  windowNames := make([]string, len(WindowFunctions), len(WindowFunctions))

  i := 0
  for windowName := range WindowFunctions {
    windowNames[i] = windowName
    i++
  }
  return windowNames
}

func WindowNamesString() string {
  return strings.Join(WindowNames(), ", ")
}

func SineTable(sineTable []float64) {
  tableLen := len(sineTable)

  for i := 0; i < tableLen; i++ {
    sineTable[i] = 0.5 * math.Cos(float64(i) * twoPi / float64(tableLen))
  }
}

// helper for hamm/hann windows
func ingWindow(windowSize int, a float64) (window []float64) {
  b := 1 - a
  window = make([]float64, windowSize, windowSize)

  for i := 0; i < windowSize; i++ {
    window[i] = a - b * math.Cos((twoPi * float64(i)) / float64(windowSize - 1))
  }

  return window
}

func HammingWindow(windowSize int) []float64 {
  return ingWindow(windowSize, 0.54)
}

// this is a more standard hann window and differs from that found in
// SoundHack Math.c:56. Soundhack sets variable b to 0.4. Normally
// b = 1 - a in standard hann windows, so b == 0.5 in our case
func VonHannWindow(windowSize int) []float64 {
  return ingWindow(windowSize, 0.5)
}

func KaiserWindow(windowSize int) []float64 {
  window := make([]float64, windowSize, windowSize)

  halfSize := windowSize / 2
  bes := besseli(6.8)
  xind :=float64((windowSize - 1) * (windowSize - 1))

  for i := 0; i < halfSize; i++ {
    floati := float64(i)
    floati = 4.0 * floati * floati
    floati = math.Sqrt(1.0 - floati / xind)
    window[i + halfSize] = besseli(6.8 * floati) / bes
    window[halfSize - i] = window[i + halfSize]
  }
  window[windowSize - 1] = 0
  window[0] = 0

  return window
}

func RampWindow(windowSize int) []float64 {
  window := make([]float64, windowSize, windowSize)

  for i := 0; i < windowSize; i++ {
    tmpFloat := float64(i) / float64(windowSize)
    window[i] = 1.0 - tmpFloat
  }

  return window
}

func RectangleWindow(windowSize int) []float64 {
  window := make([]float64, windowSize, windowSize)

  for i := 0; i < windowSize; i++ {
    window[i] = 1.0
  }

  return window
}

func SincWindow(windowSize int) []float64 {
  window := make([]float64, windowSize, windowSize)

  halfSize := float64(windowSize) / 2.0

  for i := 0; i < windowSize; i++ {
    floati := float64(i)

    if halfSize == floati {
      window[i] = 1.0;
    } else {
      window[i] = float64(windowSize) * (
        math.Sin(pi * (floati - halfSize) / halfSize) /
        (2.0 * pi * (floati - halfSize)))
    }
  }

  return window
}

func TriangleWindow(windowSize int) []float64 {
  window := make([]float64, windowSize, windowSize)

  up := true
  tmpFloat := 0.0;

  floatSize := float64(windowSize)

  for i := 0; i < windowSize; i++ {
    window[i] = 2.0 * tmpFloat
    if up {
      tmpFloat = tmpFloat + 1.0 / floatSize

      if tmpFloat > 0.5 {
        tmpFloat = 1.0 - tmpFloat
        up = false
      }
    } else {
      tmpFloat = tmpFloat - 1.0 / floatSize
    }
  }

  return window
}

func ScaleWindowsInPlace(analysisWindow, synthesisWindow []float64, points int, interpolation int) {
  windowSize := len(analysisWindow)
  pointsFloat := float64(points)
  interpolationFloat := float64(interpolation)

  /*
  Note from original SoundHack code:
  when windowSize > points, also apply sin(x)/x windows to
  ensure that window are 0 at increments of points (the FFT length)
  away from the center of the analysis window and of interpolation
  away from the center of the synthesis window
  */

  if windowSize > points {
    halfWindowSize := -(float64(windowSize) - 1.0) / 2.0
    for i := 0; i < windowSize; i, halfWindowSize = i + 1, halfWindowSize + 1.0 {
      if halfWindowSize != 0.0 {
        analysisWindow[i] = analysisWindow[i] * pointsFloat * math.Sin(pi * halfWindowSize / pointsFloat) / (pi * halfWindowSize)

        // SoundHack PhaseVocoder.c:33 has a conditional here... if (interpolation), not sure why that's needed? just a guard against division by 0?
        synthesisWindow[i] = synthesisWindow[i] * interpolationFloat * math.Sin(pi * halfWindowSize / interpolationFloat) / (pi * halfWindowSize)
      }
    }
  }

  /*
  Note from original SoundHack code:
  normalize windows for unity gain across unmodified
  analysis-synthesis procedure
  */

  sum := 0.0
  for i := 0; i < windowSize; i++ {
    sum += analysisWindow[i]
  }

  analFactor := 2.0 / sum

  // NOTE: we are ignoring the analysisType === CSOUND code branch
  // in the original SoundHack code, I do not think analysisType is
  // even set in the Pvoc routines in Soundhack
  synthFactor := analFactor

  if windowSize > points {
    synthFactor = 1.0 / analFactor
  }

  for i := 0; i < windowSize; i++ {
    analysisWindow[i] *= analFactor
    synthesisWindow[i] *= synthFactor
  }

  if windowSize <= points {
    sum = 0.0
    for i := 0; i < windowSize; i += interpolation {
      sum += synthesisWindow[i] * synthesisWindow[i]
    }

    sum = 1.0 / sum

    for i := 0; i < windowSize; i++ {
      synthesisWindow[i] *= sum
    }
  }
}

// modified bessel function of the first kind
func besseli(x float64) float64 {
  y := x / 2.0
  t := 1.e-08
  e := 1.0
  de := 1.0

  for i := 1; i <= 25; i++ {
    xi := float64(i)
    de = de * y / xi
    sde := de * de
    e += sde
    if  e * t > sde {
      break
    }
  }

  return e
}
