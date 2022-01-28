package audioio

import(
  "fmt"
  "errors"
  "os"
  "github.com/go-audio/aiff"
  "github.com/go-audio/audio"
)

var IntMaxSignedValue = map[int]int {
  8: 127,
  16: 32767,
  24: 8388607,
  32: 2147483647,
}

type AiffReader struct {
  NumChans int
  BitDepth int
  SampleRate int
  ReadBuffer *audio.IntBuffer
  Filepath string
  NumSampleFrames int
  Duration float64
  decoder *aiff.Decoder
  fileReader *os.File
}

type AiffWriter struct {
  Filepath string
  NumChans int
  BitDepth int
  SampleRate int
  WriteBuffer *audio.IntBuffer
  encoder *aiff.Encoder
  fileWriter *os.File
  maxSampleValue int
}

// bufferLength: how many frames to read at one time
func (ar *AiffReader) Open(bufferLength int) error {
  var err error

  ar.fileReader, err = os.Open(ar.Filepath)

  if err != nil {
    return err
  }

  ar.decoder = aiff.NewDecoder(ar.fileReader)

  ar.decoder.ReadInfo()

  if ar.decoder.NumChans == 0 {
    return errors.New("AiffReader.decoder.NumChans is 0")
  }

  if ar.decoder.SampleRate == 0 {
    return errors.New("AiffReader.decoder.SampleRate is 0")
  }

  if ar.decoder.BitDepth == 0 {
    return errors.New("AiffReader.decoder.BitDepth is 0")
  }

  ar.NumChans = int(ar.decoder.NumChans)
  ar.BitDepth = int(ar.decoder.BitDepth)
  ar.SampleRate = int(ar.decoder.SampleRate)
  ar.NumSampleFrames = int(ar.decoder.NumSampleFrames)
  duration, err := ar.decoder.Duration()

  if err != nil {
    return err
  }

  ar.Duration = duration.Seconds()

  format := &audio.Format{
    NumChannels: ar.NumChans,
    SampleRate: ar.SampleRate,
  }

  ar.ReadBuffer = &audio.IntBuffer{
    Format: format,
    Data: make([]int, bufferLength * ar.NumChans, bufferLength * ar.NumChans),
    SourceBitDepth: ar.BitDepth,
  }

  return nil
}

// channel is zero indexed
func (ar *AiffReader) ExtractChannel(channel int) (*audio.IntBuffer, error) {
  if ar.NumChans == 0 {
    return nil, errors.New("AiffReader.has no channels to extract")
  }

  if channel > ar.NumChans - 1 {
    return nil, fmt.Errorf("Requested channel (%d) is out of bounds 0-%d", channel, ar.NumChans - 1)
  }

  buffer := &audio.IntBuffer{
    Format: ar.ReadBuffer.Format,
    Data: make([]int, ar.ReadBuffer.NumFrames(), ar.ReadBuffer.NumFrames()),
    SourceBitDepth: ar.ReadBuffer.SourceBitDepth,
  }

  x := 0
  for i := channel; i < len(ar.ReadBuffer.Data); i += ar.NumChans {
    buffer.Data[x] = ar.ReadBuffer.Data[i]
    x++
  }

  return buffer, nil
}

func (ar *AiffReader) Close() {
  ar.fileReader.Close()
}

// numSamples is the number of samples read across all channels
// numFrames is the number of samples per channel
func (ar *AiffReader) ReadNext() (numSamples, numFrames int, err error) {
  numSamples, err = ar.decoder.PCMBuffer(ar.ReadBuffer)
  numFrames = numSamples / ar.NumChans
  return
}

// AiffWriter
func (aw *AiffWriter) Create(bufferLength int) error {
  var err error

  aw.fileWriter, err = os.Create(aw.Filepath)

  if err != nil {
    return err
  }

  aw.encoder = aiff.NewEncoder(
    aw.fileWriter,
    aw.SampleRate,
    aw.BitDepth,
    aw.NumChans,
  )

  format := &audio.Format{
    NumChannels: aw.NumChans,
    SampleRate: aw.SampleRate,
  }

  aw.WriteBuffer = &audio.IntBuffer{
    Format: format,
    Data: make([]int, bufferLength * aw.NumChans, bufferLength * aw.NumChans),
    SourceBitDepth: aw.BitDepth,
  }

  aw.maxSampleValue = IntMaxSignedValue[aw.BitDepth]

  if aw.maxSampleValue == 0 {
    return fmt.Errorf("BitDepth %d returned invalid integer max signed value of 0", aw.BitDepth)
  }

  return nil
}

func (aw *AiffWriter) Close() {
  aw.encoder.Close()
  aw.fileWriter.Close()
}

func (aw *AiffWriter) Write(buffer *audio.IntBuffer) error {
  // clip gaurd: if any sample in the int buffer exceeds maximum allowed for the
  // buffer's BitDepth, clip the sample instead of letting the encoder have it
  for i := 0; i < len(buffer.Data); i++ {
    if buffer.Data[i] > aw.maxSampleValue {
      buffer.Data[i] = aw.maxSampleValue
    } else if buffer.Data[i] < -aw.maxSampleValue {
      buffer.Data[i] = -aw.maxSampleValue
    }
  }

  return aw.encoder.Write(buffer)
}

func (aw *AiffWriter) ZeroWriteBuffer() {
  for i := 0; i < len(aw.WriteBuffer.Data); i++ {
    aw.WriteBuffer.Data[i] = 0
  }
}

func (aw *AiffWriter) WriteNext() error {
  return aw.Write(aw.WriteBuffer)
}

func (aw *AiffWriter) InterleaveChannel(channel int, data []int) error {
  if len(data) * aw.NumChans != len(aw.WriteBuffer.Data) {
    return errors.New("Data to interleave will not fit exactly into WriteBuffer")
  }

  for frameNumber :=0; frameNumber < len(data); frameNumber++ {
    i := frameNumber * aw.NumChans
    aw.WriteBuffer.Data[i + channel] = data[frameNumber]
  }

  return nil
}
