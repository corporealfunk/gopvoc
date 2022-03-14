package audioio

import(
  "fmt"
  "errors"
  "os"
  "github.com/go-audio/wav"
  "github.com/go-audio/audio"
)

type WaveReader struct {
  AudioFile
  ReadBuffer *audio.IntBuffer
  NumSampleFrames int
  Duration float64
  decoder *wav.Decoder
  fileIo *os.File
}

type WaveWriter struct {
  AudioFile
  WriteBuffer *audio.IntBuffer
  encoder *wav.Encoder
  maxSampleValue int
  fileIo *os.File
}

// Getters
func (wr *WaveReader) GetBitDepth() int {
  return wr.BitDepth
}

func (wr *WaveReader) GetSampleRate() int {
  return wr.SampleRate
}

func (wr *WaveReader) GetNumChans() int {
  return wr.NumChans
}

func (wr *WaveReader) GetNumSampleFrames() int {
  return wr.NumSampleFrames
}

func (wr *WaveReader) GetDuration() float64 {
  return wr.Duration
}

// bufferLength: how many frames to read at one time
func (wr *WaveReader) Open(bufferLength int) error {
  var err error

  wr.fileIo, err = os.Open(wr.Filepath)

  if err != nil {
    return err
  }

  wr.decoder = wav.NewDecoder(wr.fileIo)

  wr.decoder.ReadInfo()

  if wr.decoder.NumChans == 0 {
    return errors.New("WaveReader.decoder.NumChans is 0")
  }

  if wr.decoder.SampleRate == 0 {
    return errors.New("WaveReader.decoder.SampleRate is 0")
  }

  if wr.decoder.BitDepth == 0 {
    return errors.New("WaveReader.decoder.BitDepth is 0")
  }

  wr.NumChans = int(wr.decoder.NumChans)
  wr.BitDepth = int(wr.decoder.BitDepth)
  wr.SampleRate = int(wr.decoder.SampleRate)
  duration, err := wr.decoder.Duration()

  if err != nil {
    return err
  }

  wr.Duration = duration.Seconds()
  wr.NumSampleFrames = int(wr.Duration * float64(wr.SampleRate))

  format := &audio.Format{
    NumChannels: wr.NumChans,
    SampleRate: wr.SampleRate,
  }

  wr.ReadBuffer = &audio.IntBuffer{
    Format: format,
    Data: make([]int, bufferLength * wr.NumChans, bufferLength * wr.NumChans),
    SourceBitDepth: wr.BitDepth,
  }

  return nil
}

// channel is zero indexed
func (wr *WaveReader) ExtractChannel(channel int) (*audio.IntBuffer, error) {
  if wr.NumChans == 0 {
    return nil, errors.New("WaveReader.has no channels to extract")
  }

  if channel > wr.NumChans - 1 {
    return nil, fmt.Errorf("Requested channel (%d) is out of bounds 0-%d", channel, wr.NumChans - 1)
  }

  buffer := &audio.IntBuffer{
    Format: wr.ReadBuffer.Format,
    Data: make([]int, wr.ReadBuffer.NumFrames(), wr.ReadBuffer.NumFrames()),
    SourceBitDepth: wr.ReadBuffer.SourceBitDepth,
  }

  x := 0
  for i := channel; i < len(wr.ReadBuffer.Data); i += wr.NumChans {
    buffer.Data[x] = wr.ReadBuffer.Data[i]
    x++
  }

  return buffer, nil
}

func (wr *WaveReader) Close() {
  wr.fileIo.Close()
}

// numSamples is the number of samples read across all channels
// numFrames is the number of samples per channel
func (wr *WaveReader) ReadNext() (numSamples, numFrames int, err error) {
  numSamples, err = wr.decoder.PCMBuffer(wr.ReadBuffer)
  numFrames = numSamples / wr.NumChans
  return
}

// WaveWriter
func (wr *WaveWriter) Create(bufferLength int) error {
  var err error

  wr.fileIo, err = os.Create(wr.Filepath)

  if err != nil {
    return err
  }

  wr.encoder = wav.NewEncoder(
    wr.fileIo,
    wr.SampleRate,
    wr.BitDepth,
    wr.NumChans,
    1, // Linear PCM
  )

  format := &audio.Format{
    NumChannels: wr.NumChans,
    SampleRate: wr.SampleRate,
  }

  wr.WriteBuffer = &audio.IntBuffer{
    Format: format,
    Data: make([]int, bufferLength * wr.NumChans, bufferLength * wr.NumChans),
    SourceBitDepth: wr.BitDepth,
  }

  wr.maxSampleValue = IntMaxSignedValue[wr.BitDepth]

  if wr.maxSampleValue == 0 {
    return fmt.Errorf("BitDepth %d returned invalid integer max signed value of 0", wr.BitDepth)
  }

  return nil
}

func (wr *WaveWriter) Close() {
  wr.encoder.Close()
  wr.fileIo.Close()
}

func (wr *WaveWriter) Write(buffer *audio.IntBuffer) error {
  // clip gaurd: if any sample in the int buffer exceeds maximum allowed for the
  // buffer's BitDepth, clip the sample instead of letting the encoder have it
  for i := 0; i < len(buffer.Data); i++ {
    if buffer.Data[i] > wr.maxSampleValue {
      buffer.Data[i] = wr.maxSampleValue
    } else if buffer.Data[i] < -wr.maxSampleValue {
      buffer.Data[i] = -wr.maxSampleValue
    }
  }

  return wr.encoder.Write(buffer)
}

func (wr *WaveWriter) ZeroWriteBuffer() {
  for i := 0; i < len(wr.WriteBuffer.Data); i++ {
    wr.WriteBuffer.Data[i] = 0
  }
}

func (wr *WaveWriter) WriteNext() error {
  return wr.Write(wr.WriteBuffer)
}

func (wr *WaveWriter) InterleaveChannel(channel int, data []int) error {
  if len(data) * wr.NumChans != len(wr.WriteBuffer.Data) {
    return errors.New("Data to interleave will not fit exactly into WriteBuffer")
  }

  for frameNumber :=0; frameNumber < len(data); frameNumber++ {
    i := frameNumber * wr.NumChans
    wr.WriteBuffer.Data[i + channel] = data[frameNumber]
  }

  return nil
}
