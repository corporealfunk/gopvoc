package audioio

import(
  "testing"
  . "gopvoc/testing_utilities"
)

func TestReturnFileType_AIFF(t *testing.T) {
  result, _ := returnFileType("../fixtures/sine_1_chan.aif")
  Equals(
    t,
    TYPE_AIFF,
    result,
  )
}

func TestReturnFileType_WAVE(t *testing.T) {
  result, _ := returnFileType("../fixtures/sine_1_chan.wav")
  Equals(
    t,
    TYPE_WAVE,
    result,
  )
}

func TestReturnFileType_INVALID(t *testing.T) {
  result, err := returnFileType("../fixtures/textfile.txt")
  Equals(
    t,
    TYPE_INVALID,
    result,
  )

  Assert(
    t,
    err.Error() == "Invalid File Type",
    "Error was incorrect",
  )
}

func TestReturnFileTypeFromExtension_aif(t *testing.T) {
  result, _ := returnFileTypeFromExtension("../fixtures/sine_1_chan.aif")
  Equals(
    t,
    TYPE_AIFF,
    result,
  )

  result, _ = returnFileTypeFromExtension("../fixtures/sine_1_chan.aiFf")
  Equals(
    t,
    TYPE_AIFF,
    result,
  )
}

func TestReturnFileTypeFromExtension_wav(t *testing.T) {
  result, _ := returnFileTypeFromExtension("../fixtures/sine_1_chan.wav")
  Equals(
    t,
    TYPE_WAVE,
    result,
  )

  result, _ = returnFileTypeFromExtension("../fixtures/sine_1_chan.Wave")
  Equals(
    t,
    TYPE_WAVE,
    result,
  )
}

func TestReturnFileTypeFromExtension_INVALID(t *testing.T) {
  result, err := returnFileType("../fixtures/textfile.txt")
  Equals(
    t,
    TYPE_INVALID,
    result,
  )

  Assert(
    t,
    err.Error() == "Invalid File Type",
    "Error was incorrect",
  )
}
