package pvoc

import(
  "testing"
  . "gopvoc/testing_utilities"
)

// helper to setup computation of salefactor:

func returnTimeScaleData(bands, overlap, scaleFactor float64) timeScalingData {
  windowSize := int(float64(bands) * 2.0 * overlap)

  return computeTimeScaleData(
    windowSize,
    scaleFactor,
  )
}


func TestComputeTimeScaleData(t *testing.T) {
  // scaleFactor > 1
  var result timeScalingData
  result = returnTimeScaleData(
    4096,   // bands
    4.0,    // overlap
    1000.0, // scaleFactor
  )

  Equals(t, 4, result.decimation)
  Equals(t, 4039, result.interpolation)
  Equals(t, 1009.75, result.scaleFactor)

  // scaleFactor < 1
  result = returnTimeScaleData(
    4096,   // bands
    4.0,    // overlap
    0.01,   // scaleFactor
  )

  Equals(t, 4039, result.decimation)
  Equals(t, 40, result.interpolation)
  Equals(t, float64(40)/float64(4039), result.scaleFactor)
  Assert(t, !result.rateLimited, "Rate Limited should be false")


  // extreme factors > 1.0
  result = returnTimeScaleData(
    8,      // bands
    1.0,    // overlap
    10.0,   // scaleFactor
  )

  Equals(t, 1, result.decimation)
  Equals(t, 2, result.interpolation)
  Equals(t, 2.0, result.scaleFactor)
  Assert(t, result.rateLimited, "Rate Limited should be true")

  // extreme factors < 1.0
  result = returnTimeScaleData(
    8,      // bands
    1.0,    // overlap
    0.001,   // scaleFactor
  )

  Equals(t, 2, result.decimation)
  Equals(t, 1, result.interpolation)
  Equals(t, 0.5, result.scaleFactor)
  Assert(t, result.rateLimited, "Rate Limited should be true")

  // inbetween factors
  result = returnTimeScaleData(
    64,     // bands
    1.0,    // overlap
    6.7,    // scaleFactor
  )

  Equals(t, 2, result.decimation)
  Equals(t, 14, result.interpolation)
  Equals(t, 7.0, result.scaleFactor)
  Assert(t, !result.rateLimited, "Rate Limited should be false")
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
