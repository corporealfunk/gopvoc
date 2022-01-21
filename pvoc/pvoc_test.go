package pvoc

import(
  "testing"
)

func TestComputeTimeScaleData(t *testing.T) {
  // scaleFactor > 1
  bands := 4096
  overlap := 4.0
  scaleFactor := 1000.0

  windowSize := int(float64(bands) * 2.0 * overlap)

  result := computeTimeScaleData(
    windowSize,
    scaleFactor,
  )

  if result.decimation != 4 {
    t.Errorf("> 1. decimation got %d, expected %d", result.decimation, 4)
  }

  if result.interpolation != 4039 {
    t.Errorf("> 1. interpolation got %d, expected %d", result.interpolation, 4039)
  }

  if result.scaleFactor != 1009.75 {
    t.Errorf("> 1. scaleFactor got %f, expected %f", result.scaleFactor, 1009.75)
  }

  // scaleFactor < 1
  bands = 4096
  overlap = 4.0
  scaleFactor = 0.01

  windowSize = int(float64(bands) * 2.0 * overlap)

  result = computeTimeScaleData(
    windowSize,
    scaleFactor,
  )

  if result.decimation != 4039 {
    t.Errorf("> 1. decimation got %d, expected %d", result.decimation, 4039)
  }

  if result.interpolation != 40 {
    t.Errorf("> 1. interpolation got %d, expected %d", result.interpolation, 40)
  }

  expectedScaleFactor := float64(40)/float64(4039)
  if result.scaleFactor !=  expectedScaleFactor {
    t.Errorf("> 1. scaleFactor got %f, expected %f", result.scaleFactor, expectedScaleFactor)
  }
}

func equalSlices(a, b []float64) bool {
  if len(a) != len(b) {
    return false
  }
  for i, v := range a {
    if v != b[i] {
      return false
    }
  }
  return true
}

func TestSlidingBufferShiftIn(t *testing.T) {
  slidingBuffer := NewSlidingBuffer(5)

  // init
  slidingBuffer.ShiftIn([]float64{1.0, 2.0, 3.0, 4.0, 5.0}, 5)

  slidingBuffer.ShiftIn([]float64{50.0, 60.0}, 2)

  // call 1
  if !equalSlices(slidingBuffer.Data, []float64{3.0, 4.0, 5.0, 50.0, 60.0}) {
    t.Errorf("SlidingBuffer shiftIn 1 result unexpected %f", slidingBuffer.Data)
  }

  if slidingBuffer.lastValidSample != 4 {
    t.Errorf("SlidingBuffer shiftIn 1 last valid Sample unexpected: %d", slidingBuffer.lastValidSample)
  }

  // call 2
  slidingBuffer.ShiftIn([]float64{10.0, 20.0, 30.0, 40.0, 50.0}, 5)
  if !equalSlices(slidingBuffer.Data, []float64{10.0, 20.0, 30.0, 40.0, 50.0}) {
    t.Errorf("SlidingBuffer shiftIn 2 result unexpected %f", slidingBuffer.Data)
  }

  if slidingBuffer.lastValidSample != 4 {
    t.Errorf("SlidingBuffer shiftIn 2 last valid Sample unexpected: %d", slidingBuffer.lastValidSample)
  }

  // call 3, test last valid sample
  slidingBuffer.ShiftIn([]float64{100.0, 101.0, 102.0}, 2)
  if !equalSlices(slidingBuffer.Data, []float64{40.0, 50.0, 100.0, 101.0, 0}) {
    t.Errorf("SlidingBuffer shiftIn 3 result unexpected %f", slidingBuffer.Data)
  }

  if slidingBuffer.lastValidSample != 3 {
    t.Errorf("SlidingBuffer shiftIn 3 last valid Sample unexpected: %d", slidingBuffer.lastValidSample)
  }
}

func TestSlidingBufferShiftOver(t *testing.T) {
  slidingBuffer := NewSlidingBuffer(5)

  // init
  slidingBuffer.ShiftIn([]float64{1.0, 2.0, 3.0, 4.0, 5.0}, 5)

  slidingBuffer.ShiftOver(2)

  // call 1
  if !equalSlices(slidingBuffer.Data, []float64{3.0, 4.0, 5.0, 0.0, 0}) {
    t.Errorf("SlidingBuffer shiftOver 1 result unexpected %f", slidingBuffer.Data)
  }

  if slidingBuffer.lastValidSample != 2 {
    t.Errorf("SlidingBuffer shiftOver 1 last valid Sample unexpected: %d", slidingBuffer.lastValidSample)
  }

  // call 2
  slidingBuffer.ShiftOver(3)
  if !equalSlices(slidingBuffer.Data, []float64{0, 0, 0, 0, 0}) {
    t.Errorf("SlidingBuffer shiftOver 2 result unexpected %f", slidingBuffer.Data)
  }

  if slidingBuffer.lastValidSample != -1 {
    t.Errorf("SlidingBuffer shiftOver 2 last valid Sample unexpected: %d", slidingBuffer.lastValidSample)
  }
}
