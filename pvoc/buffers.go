package pvoc

import(
  "fmt"
  "math"
)

type SlidingBuffer struct {
  Data []float64
  lastValidSample int
  hasReceivedData bool
}

func NewSlidingBuffer(length int) (buffer *SlidingBuffer) {
  buffer = &SlidingBuffer{
    Data: make([]float64, length, length),
    lastValidSample: -1,
    hasReceivedData: false,
  }

  return buffer
}

func (sb *SlidingBuffer) HasValidSamples() bool {
  return sb.lastValidSample >= 0
}

func (sb *SlidingBuffer) ShiftIn(data []float64, validSamples int) error {
  dataLen := len(data)

  if validSamples > dataLen {
    return fmt.Errorf("validSamples %d cannot be more than buffer data length %d", validSamples, dataLen)
  }

  if dataLen > len(sb.Data) {
    return fmt.Errorf("Attempted to ShiftIn %d samples, but buffer can only hold %d samples", dataLen, len(sb.Data))
  }

  // shift the data over by len(data), then copy the data in
  for i := dataLen; i < len(sb.Data); i++ {
    sb.Data[i - dataLen] = sb.Data[i]
  }

  // copy new data to end
  for i := 0; i < dataLen; i++ {
    sb.Data[len(sb.Data) - dataLen + i] = data[i]
  }

  // pad with zeros any non-valid samples
  if !sb.hasReceivedData {
    sb.lastValidSample = len(sb.Data) - dataLen + validSamples - 1
  } else {
    sb.lastValidSample = sb.lastValidSample - dataLen + validSamples
  }

  for i := len(sb.Data) - dataLen + validSamples; i < len(sb.Data); i++ {
    sb.Data[i] = 0
  }

  sb.hasReceivedData = true

  return nil
}

// shifts over by a given length, pads the rest with 0s
func (sb *SlidingBuffer) ShiftOver(dataLen int) error {
  if dataLen > len(sb.Data) {
    return fmt.Errorf("Attempted to ShiftOver %d samples, but buffer can only hold %d samples", dataLen, len(sb.Data))
  }

  // shift the data over by dataLen
  for i := dataLen; i < len(sb.Data); i += 1 {
    sb.Data[i - dataLen] = sb.Data[i]
  }

  // pad with 0s
  for i := len(sb.Data) - dataLen; i < len(sb.Data); i += 1 {
    sb.Data[i] = 0
  }

  if !sb.hasReceivedData {
    sb.lastValidSample = len(sb.Data) - dataLen - 1
  } else {
    sb.lastValidSample = sb.lastValidSample - dataLen
  }

  sb.hasReceivedData = true

  return nil
}

func (sb *SlidingBuffer) DataInts() (buffer []int) {
  buffer = make([]int, len(sb.Data), len(sb.Data))

  for i := 0; i < len(sb.Data); i++ {
    buffer[i] = int(math.Round(sb.Data[i]))
  }

  return buffer
}
