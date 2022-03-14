package audioio

import(
  "os"
  "fmt"
  "github.com/go-audio/audio"
  "bytes"
  "path/filepath"
  "strings"
)

var IntMaxSignedValue = map[int]int {
  8: 127,
  16: 32767,
  24: 8388607,
  32: 2147483647,
}

const TYPE_INVALID = -1
const TYPE_AIFF = 1
const TYPE_WAVE = 2

type Reader interface {
  Open(bufferLength int) error
  Close()
  ReadNext() (int, int, error)
  ExtractChannel(channel int) (*audio.IntBuffer, error)
  GetBitDepth() int
  GetSampleRate() int
  GetNumChans() int
  GetNumSampleFrames() int
  GetDuration() float64
}

type Writer interface {
  Create(bufferLength int) error
  Close()
  Write(buffer *audio.IntBuffer) error
  WriteNext() error
  InterleaveChannel(channel int, data []int) error
  ZeroWriteBuffer()
}

type AudioFile struct {
  Filepath string
  NumChans int
  BitDepth int
  SampleRate int
}

type AudioReader struct {
  Reader Reader
  fileType int
}

type AudioWriter struct {
  Writer Writer
  fileType int
}

// determines a filetype based on the given file extension, the file does not have to exist
func returnFileTypeFromExtension(filePath string) (int, error) {
  extension := strings.ToLower(filepath.Ext(filePath))

  switch extension {
  case ".aiff":
    return TYPE_AIFF, nil
  case ".aif":
    return TYPE_AIFF, nil
  case ".wave":
    return TYPE_WAVE, nil
  case ".wav":
    return TYPE_WAVE, nil
  }

  return TYPE_INVALID, fmt.Errorf("Invalid File Type")
}

// Reades the magic bytes of the given file and returns the file type const.
// File must exist on disk
func returnFileType(filePath string) (int, error) {
  // determine the filetype
  file, err := os.Open(filePath)

  if err != nil {
    return TYPE_INVALID, err
  }

  defer file.Close()

  headerBytes := make([]byte, 12)
  if _, err := file.Read(headerBytes); err != nil {
    return TYPE_INVALID, err
  }
  headerBytes8 := []byte{}
  headerBytes8 = append(headerBytes8, headerBytes[:4]...)
  headerBytes8 = append(headerBytes8, headerBytes[8:]...)

  if bytes.Equal(headerBytes8, []byte("FORMAIFF")) {
    return TYPE_AIFF, nil
  } else if bytes.Equal(headerBytes8, []byte("RIFFWAVE")) {
    return TYPE_WAVE, nil
  }

  return TYPE_INVALID, fmt.Errorf("Invalid File Type")
}

func NewAudioReader(filePath string) (ar *AudioReader, err error) {
  ar = &AudioReader{}

  // get file type
  fileType, err := returnFileType(filePath)

  if err != nil {
    return nil, err
  }

  switch fileType {
  case TYPE_AIFF:
    audioFile := AudioFile{Filepath: filePath}
    ar.Reader = &AiffReader{AudioFile: audioFile}
    ar.fileType = TYPE_AIFF
  case TYPE_WAVE:
    audioFile := AudioFile{Filepath: filePath}
    ar.Reader = &WaveReader{AudioFile: audioFile}
    ar.fileType = TYPE_WAVE
  default:
    return nil, fmt.Errorf("AudioReader doesn't implement filetype %d", fileType)
  }

  return ar, nil
}

// delegate to the reader
func (ar *AudioReader) Open(bufferLength int) (err error) {
  return ar.Reader.Open(bufferLength)
}

func (ar *AudioReader) Close() {
  ar.Reader.Close()
}

func (ar *AudioReader) ReadNext() (int, int, error) {
  return ar.Reader.ReadNext()
}

func (ar *AudioReader) ExtractChannel(channel int) (*audio.IntBuffer, error) {
  return ar.Reader.ExtractChannel(channel)
}

func (ar *AudioReader) GetNumChans() int {
  return ar.Reader.GetNumChans()
}

func (ar *AudioReader) GetBitDepth() int {
  return ar.Reader.GetBitDepth()
}

func (ar *AudioReader) GetSampleRate() int {
  return ar.Reader.GetSampleRate()
}

func (ar *AudioReader) GetNumSampleFrames() int {
  return ar.Reader.GetNumSampleFrames()
}

func (ar *AudioReader) GetDuration() float64 {
  return ar.Reader.GetDuration()
}

// Audio Writer
func NewAudioWriter(audioFile AudioFile) (aw *AudioWriter, err error) {
  aw = &AudioWriter{}

  // get file type
  fileType, err := returnFileTypeFromExtension(audioFile.Filepath)

  if err != nil {
    return nil, err
  }

  switch fileType {
  case TYPE_AIFF:
    aw.Writer = &AiffWriter{AudioFile: audioFile}
    aw.fileType = TYPE_AIFF
  case TYPE_WAVE:
    aw.Writer = &WaveWriter{AudioFile: audioFile}
    aw.fileType = TYPE_WAVE
  default:
    return nil, fmt.Errorf("AudioWriter doesn't implement filetype %d", fileType)
  }

  return aw, nil
}

// delegate to Writer
func (aw *AudioWriter) Create(bufferLength int) error {
  return aw.Writer.Create(bufferLength)
}

func (aw *AudioWriter) Close() {
  aw.Writer.Close()
}

func (aw *AudioWriter) ZeroWriteBuffer() {
  aw.Writer.ZeroWriteBuffer()
}

func (aw *AudioWriter) InterleaveChannel(channel int, data []int) error {
  return aw.Writer.InterleaveChannel(channel, data)
}

func (aw *AudioWriter) WriteNext() error {
  return aw.Writer.WriteNext()
}
