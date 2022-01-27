package pvoc

import(
  "fmt"
  "math"
  "gopvoc/audioio"
  // "gopvoc/charter"
)

// FFT directions
const Time2Freq = 1
const Freq2Time = 2

// Processing operations
const TimeStretch = 3
const PitchShift = 4

var OperationNames = map[int]string {
  TimeStretch: "Time Scale",
  PitchShift: "Pitch Shift",
}

var allowedOverlaps = map[float64]bool {
  0.5: true,
  1.0: true,
  2.0: true,
  4.0: true,
}

type timeScalingData struct {
  scaleFactor float64
  decimation int
  interpolation int
  rateLimited bool
}

type Pvoc struct {
  Bands int
  Overlap float64
  ScaleFactor float64
  Points int
  WindowSize int
  Decimation int
  Interpolation int
  Operation int
  PhaseLock bool // only useful for TimeStretch
  WindowName string
  GatingAmplitudeDb float64
  GatingThresholdDb float64
  RateLimited bool // only set for TimeStretch
  gatingAmplitude float64
  gatingThreshold float64
}

func NewPvoc(
  bands int,
  overlap,
  scaleFactor float64,
  operation int,
  phaseLock bool,
  windowName string,
  gatingAmplitudeDb,
  gatingThresholdDb float64,
) (*Pvoc, error) {
  if bands > 4096 || bands < 1 || (bands & (bands - 1)) != 0 {
    return nil, fmt.Errorf("bands must be a power of 2 less than or equal to 4096, got %d", bands)
  }

  if !allowedOverlaps[overlap] {
    return nil, fmt.Errorf("overlap must be 0.5, 1.0, 2.0 or 4.0, got %f", overlap)
  }

  if operation != TimeStretch && operation != PitchShift {
    return nil, fmt.Errorf("Operation must be either TimeStretch (%d) or PitchShift (%d), got %d", TimeStretch, PitchShift, operation)
  }

  if scaleFactor < 0 {
    return nil, fmt.Errorf("Scale multiplier cannot be negative, got %f", scaleFactor)
  }

  if gatingAmplitudeDb > 0 {
    return nil, fmt.Errorf("Resynthesis gating amplitude must be less than 0, got %f.", gatingAmplitudeDb)
  }

  if gatingThresholdDb > 0 {
    return nil, fmt.Errorf("Resynthesis gating threshold below maximum must be less than 0, got %f.", gatingThresholdDb)
  }

  gatingAmplitude := 0.0
  if gatingAmplitudeDb != 0 {
    gatingAmplitude = math.Pow(10.0, gatingAmplitudeDb / 20.0)
  }

  gatingThreshold := 0.0
  if gatingThresholdDb != 0 {
    gatingThreshold = math.Pow(10.0, gatingThresholdDb / 20.0)
  }

  pvoc := &Pvoc{
    Bands: bands,
    Overlap: overlap,
    ScaleFactor: scaleFactor,
    Points: bands * 2,
    WindowSize: int(float64(bands) * 2.0 * overlap),
    Operation: operation,
    PhaseLock: phaseLock,
    WindowName: windowName,
    GatingAmplitudeDb: gatingAmplitudeDb,
    GatingThresholdDb: gatingThresholdDb,
    gatingAmplitude: gatingAmplitude,
    gatingThreshold: gatingThreshold,
  }

  if operation == TimeStretch {
    timeScalingData := computeTimeScaleData(pvoc.WindowSize, pvoc.ScaleFactor)

    pvoc.ScaleFactor = timeScalingData.scaleFactor
    pvoc.Interpolation = timeScalingData.interpolation
    pvoc.Decimation = timeScalingData.decimation
    pvoc.RateLimited = timeScalingData.rateLimited
  } else {
    pvoc.ScaleFactor = scaleFactor
    pvoc.Interpolation = int(float64(bands) * overlap / 4.0)
    pvoc.Decimation = pvoc.Interpolation
  }

  return pvoc, nil
}

func (p *Pvoc) String() (output string) {
  output += fmt.Sprintf("%24s   %s\n", "Operation:", OperationNames[p.Operation])
  output += fmt.Sprintf("%24s   %d\n", "Bands:", p.Bands)
  output += fmt.Sprintf("%24s   %f\n", "Overlap:", p.Overlap)
  output += fmt.Sprintf("%24s   %f", "Scaling:", p.ScaleFactor)

  if p.Operation == TimeStretch && p.RateLimited {
    output += " (limited to "
    if p.ScaleFactor < 1.0 {
      output += "min"
    } else {
      output += "max"
    }
    output += ")"
  }
  output += "\n"
  output += fmt.Sprintf("%24s   %s\n", "Windowing Func:", p.WindowName)
  output += fmt.Sprintf("%24s   %d samples\n", "Decimation Length:", p.Decimation)
  output += fmt.Sprintf("%24s   %d samples\n", "Interpolation Length:", p.Interpolation)

  if p.Operation == TimeStretch {
    output += fmt.Sprintf("%24s   %t\n", "Phase Locking:", p.PhaseLock)
  }

  if p.GatingAmplitudeDb != 0 {
    output += fmt.Sprintf("%24s   %f\n", "Gating Amp Min:", p.GatingAmplitudeDb)
  }

  if p.GatingThresholdDb != 0 {
    output += fmt.Sprintf("%24s   %f\n", "Gating Amp Thresh <Max:", p.GatingThresholdDb)
  }
  return
}

func computeTimeScaleData(windowSize int, scaleFactor float64) timeScalingData {
  var maxRate int = windowSize / 8

  var decimation int
  var interpolation int

  minPercentError := -1.0
  bestInterpolation := 0
  bestDecimation := 0
  newScaleFactor := 1.0

  if scaleFactor > 1.0 {
    for interpolation = maxRate; interpolation > 0; interpolation-- {
      decimation = int(math.Floor(float64(interpolation) / scaleFactor))

      if decimation == 0 {
        decimation = 1
      }

      tempScaleFactor := float64(interpolation) / float64(decimation)

      var percentError float64

      if tempScaleFactor > scaleFactor {
        percentError = tempScaleFactor / scaleFactor
      } else {
        percentError = scaleFactor / tempScaleFactor
      }

      if minPercentError < 0.0 || percentError < minPercentError {
        minPercentError = percentError
        bestDecimation = decimation
        bestInterpolation = interpolation
        newScaleFactor = tempScaleFactor

        // good enough, don't compute more
        if percentError < 1.01 {
          break
        }
      }
    }
  } else {
    for decimation = maxRate; decimation > 0; decimation-- {
      interpolation = int(math.Floor(float64(decimation) * scaleFactor))

      if interpolation == 0 {
        interpolation = 1
      }

      tempScaleFactor := float64(interpolation) / float64(decimation)

      var percentError float64

      if tempScaleFactor > scaleFactor {
        percentError = tempScaleFactor / scaleFactor
      } else {
        percentError = scaleFactor / tempScaleFactor
      }

      if minPercentError < 0.0 || percentError < minPercentError {
        minPercentError = percentError
        bestDecimation = decimation
        bestInterpolation = interpolation
        newScaleFactor = tempScaleFactor

        // good enough, don't compute more
        if percentError < 1.01 {
          break
        }
      }
    }
  }

  // we are rateLimited if we asked for a larger scale factor than is possible:
  // the max scaleFactor is  computed if interp or decimation is as at MaxRate and
  // that has caused the other factor to go to 1

  rateLimited := false

  if scaleFactor > 1.0 && bestInterpolation == maxRate {
    maxScaleFactor := float64(maxRate) / 1.0
    rateLimited = scaleFactor > maxScaleFactor
  } else if scaleFactor < 1.0 && bestDecimation == maxRate {
    maxScaleFactor := 1.0 / float64(maxRate)
    rateLimited = scaleFactor < maxScaleFactor
  }

  return timeScalingData{
    scaleFactor: newScaleFactor,
    decimation: bestDecimation,
    interpolation: bestInterpolation,
    rateLimited: rateLimited,
  }
}

func (p *Pvoc) Run(
  aiffReader *audioio.AiffReader,
  aiffWriter *audioio.AiffWriter,
  progress chan<- int,
  errors chan<- error,
  done chan<- bool,
) {
  // setup the buffers for input and output
  inputBuffers := make([]*SlidingBuffer, aiffReader.NumChans, aiffReader.NumChans)
  outputBuffers := make([]*SlidingBuffer, aiffReader.NumChans, aiffReader.NumChans)

  // setup the FFT processing buffers
  spectrumBuffers := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)
  polarBuffers := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)

  // setup storage of last processed phases for interoplation for TimeStrech
  lastPhaseIns := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)
  lastPhaseOuts := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)

  // setup amp, freq and sine index storage for PitchShift and sineTable
  lastAmps := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)
  lastFreqs := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)
  sineIndexes := make([][]float64, aiffReader.NumChans, aiffReader.NumChans)
  sineTable := make([]float64, 8192, 8192)
  SineTable(sineTable)

  halfPoints := p.Points / 2

  // what is the maximum ABS sample value at our BitDepth?
  maxSampleValue := math.Pow(2, float64(aiffReader.BitDepth - 1))

  for c := 0; c < aiffReader.NumChans; c++ {
    inputBuffers[c] = NewSlidingBuffer(p.WindowSize)
    outputBuffers[c] = NewSlidingBuffer(p.WindowSize)
    spectrumBuffers[c] = make([]float64, p.Points, p.Points)
    polarBuffers[c] = make([]float64, p.Points + 2, p.Points + 2)

    lastPhaseIns[c] = make([]float64, halfPoints + 1, halfPoints + 1)

    // TimeStretch needs
    lastPhaseOuts[c] = make([]float64, halfPoints + 1, halfPoints + 1)

    // PitchShift needs
    lastAmps[c] = make([]float64, halfPoints + 1, halfPoints + 1)
    lastFreqs[c] = make([]float64, halfPoints + 1, halfPoints + 1)
    sineIndexes[c] = make([]float64, halfPoints + 1, halfPoints + 1)
  }

  // setup analysis and synthesis windows
  windowFunction := WindowFunctions[p.WindowName]

  if windowFunction == nil {
    errors <- fmt.Errorf("Invalid window function (%s), valid options are: %s", p.WindowName, WindowNamesString())
    return
  }

  analysisWindow := windowFunction(
    p.WindowSize,
  )
  synthesisWindow := windowFunction(
    p.WindowSize,
  )

  // charter.MakeChart("triangle", 1, analysisWindow)

  // scale the windows in place
  ScaleWindowsInPlace(
    analysisWindow,
    synthesisWindow,
    p.Points,
    p.Interpolation,
  )

  // where we are in the input/output in samples
  inPointer := p.WindowSize * -1
  outPointer := (inPointer * p.Interpolation) / p.Decimation

  blockCount := 0
  totalSamplesRead := 0
  progress <- 0
  for {
    inPointer += p.Decimation
    outPointer += p.Interpolation

    _, samplesRead, err := aiffReader.ReadNext()
    totalSamplesRead += samplesRead

    if err != nil {
      errors <- err
      return
    }

    // for each channel shift into the input buffers the number of samples read
    if samplesRead > 0 {
      for c := 0; c < aiffReader.NumChans; c++ {
        // always returns an audio.IntBuffer of decimation length
        channelBuffer, err := aiffReader.ExtractChannel(c)

        if err != nil {
          errors <- err
          return
        }

        err = inputBuffers[c].ShiftIn(
          channelBuffer.AsFloatBuffer().Data,
          samplesRead,
        )

        if err != nil {
          errors <- err
          return
        }
      }
    } else {
      // we've hit or passed EOF on aiffReader, slide it over anyway:
      for c := 0; c < aiffReader.NumChans; c++ {
        inputBuffers[c].ShiftOver(p.Decimation)
      }
    }

    for c := 0; c < aiffReader.NumChans; c++ {
      // fold the inputBuffers into the spectrum buffers
      WindowFold(
        inputBuffers[c].Data,
        analysisWindow,
        spectrumBuffers[c],
        inPointer,
      )

      // take time to frequency FFT in-place of the spectrum buffer
      RealFFT(spectrumBuffers[c], Time2Freq)

      // convert the FFT spectrum to polar
      CartToPolar(spectrumBuffers[c], polarBuffers[c])

      // gate the spectrum if need to
      if p.gatingAmplitude != 0.0 || p.gatingThreshold != 0.0 {
        SimpleSpectralGate(
          polarBuffers[c],
          p.Points,
          p.gatingAmplitude,
          p.gatingThreshold,
          maxSampleValue,
        )
      }

      if p.Operation == TimeStretch {
        // TimeStrech operations:
        PhaseInterpolate(
          polarBuffers[c],
          lastPhaseIns[c],
          lastPhaseOuts[c],
          p.Points,
          p.Decimation,
          p.ScaleFactor,
          p.PhaseLock, // this is always false in SoundHack
        )

        // convert the polar FFT result to cart
        PolarToCart(polarBuffers[c], spectrumBuffers[c])
        RealFFT(spectrumBuffers[c], Freq2Time)

        OverlapAdd(
          spectrumBuffers[c],
          synthesisWindow,
          outputBuffers[c].Data,
          outPointer,
        )
      } else {
        // PitchShift operations:
        AddSynth(
          polarBuffers[c],
          outputBuffers[c].Data,
          lastAmps[c],
          lastFreqs[c],
          lastPhaseIns[c],
          sineTable,
          sineIndexes[c],
          p.ScaleFactor,
          p.Interpolation,
          p.Decimation,
          p.Points,
        )
      }
    }

    // write to disk
    var checkTime int

    if p.Operation == TimeStretch {
      checkTime = outPointer + p.Interpolation
    } else {
      checkTime = outPointer + p.WindowSize - p.Interpolation
    }

    if checkTime >= 0 {
      aiffWriter.ZeroWriteBuffer()

      for c := 0; c < aiffReader.NumChans; c++ {
        err = aiffWriter.InterleaveChannel(
          c,
          outputBuffers[c].DataInts()[:p.Interpolation],
        )

        // charter.MakeChart(fmt.Sprintf("interleave_chan-%d", c), blockCount, outputBuffers[c].Data)

        if err != nil {
          errors <- err
          return
        }
      }

      // charter.MakeChart("writeBuffer", blockCount, aiffWriter.WriteBuffer.AsFloatBuffer().Data)

      if err = aiffWriter.WriteNext(); err != nil {
        errors <- err
        return
      }
    }

    // shift output buffers over by interpolation
    for c := 0; c < aiffReader.NumChans; c++ {
      outputBuffers[c].ShiftOver(p.Interpolation)
    }

    // Soundhack terminates when no more samples are read, we do this:
    // if the first channel input buffer has no more valid samples, break
    if !inputBuffers[0].HasValidSamples() {
      break
    }

    blockCount++;
    progress <- int((float64(totalSamplesRead) / float64(aiffReader.NumSampleFrames)) * 100.0)
  }
  done <- true
}

/* Comment from original SoundHack Code with our param names:
 * multiply current input Input by window Window (both of length lengthWindow);
 * using modulus arithmetic, fold and rotate windowed input into output array
 * output of (FFT) length lengthFFT according to current input time inPointer
 */
func WindowFold(inputBuffer []float64, analysisWindow []float64, spectrumBuffer []float64, inPointer int) {
  points := len(spectrumBuffer)
  windowSize := len(inputBuffer)

  // zero the spectrum buffer
  for i := 0; i < points; i++ {
    spectrumBuffer[i] = 0.0
  }

  for inPointer < 0 {
    inPointer += points
  }

  inPointer %= points

  for i := 0; i < windowSize; i++ {
    spectrumBuffer[inPointer] += inputBuffer[i] * analysisWindow[i]
    inPointer++
    if (inPointer == points) {
      inPointer = 0
    }
  }
}

/* SoundHack comment:
 * spectrum is a spectrum in RealFFT format, index.e., it contains lengthFFT real values
 * arranged as real followed by imaginary values, except for first two values, which
 * are real parts of 0 and Nyquist frequencies; convert first changes these into
 * lengthFFT/2+1 PAIRS of magnitude and phase values to be stored in output array
 * polarSpectrum.
 */
func CartToPolar(spectrum, polarSpectrum []float64) {
  points := len(spectrum)
  halfPoints := points / 2

  // zero the polarSpectrum
  for i := 0; i < len(polarSpectrum); i++ {
    polarSpectrum[i] = 0.0
  }

  /* SoundHack Comment:
   * unravel RealFFT-format spectrum: note that halfLengthFFT+1 pairs of values are produced
   */
  for bandNumber := 0; bandNumber <= halfPoints; bandNumber++ {
    realIndex := bandNumber * 2
    ampIndex := realIndex

    imagIndex := realIndex + 1;
    phaseIndex := imagIndex

    var realPart float64
    var imagPart float64

    if bandNumber == 0 {
      realPart = spectrum[realIndex]
      imagPart = 0.0
    } else if bandNumber == halfPoints {
      realPart = spectrum[1]
      imagPart = 0.0
    } else {
      realPart = spectrum[realIndex]
      imagPart = spectrum[imagIndex]
    }

    /* SoundHack comment:
     * compute magnitude & phase value from real and imaginary parts
     */
    polarSpectrum[ampIndex] = math.Hypot(realPart, imagPart)

    if polarSpectrum[ampIndex] == 0.0 {
      polarSpectrum[phaseIndex] = 0.0
    } else {
      polarSpectrum[phaseIndex] = -math.Atan2(imagPart, realPart)
    }
  }
}


/* SoundHack comment:
 * PolarToCart turns halfLengthFFT+1 PAIRS of amplitude and phase values in
 * polarSpectrum into halfLengthFFT PAIR of complex spectrum data (in RealFFT format)
 * in output array spectrum.
 */
func PolarToCart(polarSpectrum, spectrum []float64) {
  points := len(spectrum)
  halfPoints := points / 2

  for bandNumber := 0; bandNumber <= halfPoints; bandNumber++ {
    realIndex := bandNumber * 2
    ampIndex := realIndex

    imagIndex := realIndex + 1
    phaseIndex := imagIndex

    var realValue, imagValue float64

    if polarSpectrum[ampIndex] == 0.0 {
      realValue = 0.0
      imagValue = 0.0
    } else if bandNumber == 0 || bandNumber == halfPoints {
      realValue = polarSpectrum[ampIndex] * math.Cos(polarSpectrum[phaseIndex])
      imagValue = 0.0
    } else {
      realValue = polarSpectrum[ampIndex] * math.Cos(polarSpectrum[phaseIndex])
      imagValue = -polarSpectrum[ampIndex] * math.Sin(polarSpectrum[phaseIndex])
    }

    if bandNumber == halfPoints {
      realIndex = 1
    }

    spectrum[realIndex] = realValue

    if bandNumber != halfPoints && bandNumber != 0 {
      spectrum[imagIndex] = imagValue
    }
  }
}

func SimpleSpectralGate(
  polarSpectrum []float64,
  points int,
  minAmplitude,
  maskRatio,
  maxSampleValue float64,
) {
  halfPoints := points / 2

  maxAmplitude := 0.0

  /* Find maximum amplitude */
  for bandNumber := 0; bandNumber <= halfPoints; bandNumber++ {
    ampIndex := bandNumber * 2

    if polarSpectrum[ampIndex] > maxAmplitude {
      maxAmplitude = polarSpectrum[ampIndex]
    }
  }

  maskAmplitude := maskRatio * maxAmplitude

  for bandNumber := 0; bandNumber <= halfPoints; bandNumber++ {
    ampIndex := bandNumber * 2

    normalizedAmp := polarSpectrum[ampIndex] / maxSampleValue

    /* Set for Ducking */
    if polarSpectrum[ampIndex] < maskAmplitude || normalizedAmp < minAmplitude {
      polarSpectrum[ampIndex] = 0.0
    }
  }
}

func PhaseInterpolate(
  polarSpectrum,
  lastPhaseIn,
  lastPhaseOut []float64,
  points,
  decimation int,
  scaleFactor float64,
  phaseLock bool,
) {
  phasePerBand := (float64(decimation) * twoPi) / float64(points)
  halfPoints := points / 2

  for bandNumber := 0; bandNumber <= halfPoints; bandNumber++ {
    ampIndex := bandNumber * 2
    phaseIndex := ampIndex + 1
    var phaseDifference float64

    /* SoundHack comment:
     * take difference between the current phase value and previous value for each channel
     */
    if polarSpectrum[ampIndex] == 0.0 {
      // phaseDifference = 0.0; // unused, but declared in SoundHack
      polarSpectrum[phaseIndex] = lastPhaseOut[bandNumber]
    } else {
      if phaseLock {
        maxAmplitude := 0.0

        // set low band info
        if bandNumber > 1 {
          maxAmplitude = polarSpectrum[ampIndex - 2]
          phaseDifference = (polarSpectrum[phaseIndex - 2] - lastPhaseIn[bandNumber - 1]) - phasePerBand
        }

        if polarSpectrum[ampIndex] > maxAmplitude {
          maxAmplitude = polarSpectrum[ampIndex]
          phaseDifference = polarSpectrum[phaseIndex] - lastPhaseIn[bandNumber]
        }

        if bandNumber != halfPoints {
          if polarSpectrum[ampIndex + 2] > maxAmplitude {
            phaseDifference = (polarSpectrum[phaseIndex + 2] - lastPhaseIn[bandNumber + 1]) + phasePerBand
          }
        }
      } else {
        phaseDifference = polarSpectrum[phaseIndex] - lastPhaseIn[bandNumber]
      }

      lastPhaseIn[bandNumber] = polarSpectrum[phaseIndex]

      /* SoundHack comment:
       * unwrap phase differences
       *
       * while (phaseDifference > Pi)
       * phaseDifference -= twoPi;
       * while (phaseDifference < -Pi)
       * phaseDifference += twoPi;
       */

      phaseDifference *= scaleFactor
      /*
       * create new phase from interpolate/decimate ratio
       */
      polarSpectrum[phaseIndex] = lastPhaseOut[bandNumber] + phaseDifference

      for polarSpectrum[phaseIndex] > pi {
        polarSpectrum[phaseIndex] -= twoPi
      }

      for polarSpectrum[phaseIndex] < -pi {
        polarSpectrum[phaseIndex] += twoPi
      }

      lastPhaseOut[bandNumber] = polarSpectrum[phaseIndex]

      // SoundHack delcares these but does not use them:
      // phase := polarSpectrum[phaseIndex]
      // amplitude := polarSpectrum[ampIndex]
    }
  }
}

/* SoundHack comment:
 * Input are folded samples of length points; output and
 * synthesisWindow are of length lengthWindow--overlap-add windowed,
 * unrotated, unfolded input data into output
 */
func OverlapAdd(spectrum, synthesisWindow, output []float64, outPointer int) {
  points := len(spectrum)
  windowSize := len(synthesisWindow)

  for outPointer < 0 {
    outPointer += points
  }

  outPointer %= points

  for i := 0; i < windowSize; i++ {
    output[i] += spectrum[outPointer] * synthesisWindow[i]

    outPointer++
    if outPointer == points {
      outPointer = 0
    }
  }
}

/* SoundHack Comment:
 * oscillator bank resynthesizer for phase vocoder analyzer
 * uses sum of halfPoints+1 cosinusoidal table lookup oscillators to compute
 * interpolation samples of output from halfPoints+1 amplitude and phase value-pairs
 * in polarSpectrum; frequencies are scaled by scaleFactor
 */
 func AddSynth(
   polarSpectrum,
   output,
   lastAmp,
   lastFreq,
   lastPhaseIn,
   sineTable,
   sineIndex []float64,
   scaleFactor float64,
   interpolation,
   decimation,
   points int,
 ) {
   halfPoints := points / 2

   oneOvrInterp := 1.0 / float64(interpolation)
   cyclesBand := scaleFactor * 8192.0 / float64(points)
   cyclesFrame := scaleFactor * 8192.0 / (float64(decimation) * twoPi)

   var numberPartials int

   if scaleFactor > 1.0 {
     // TODO: this is truncating a float into int, does C do this?
     numberPartials = int(float64(halfPoints) / scaleFactor)
   } else {
     numberPartials = halfPoints
   }

   /* SoundHack comment:
   * convert phase representation into instantaneous frequency- this makes polarSpectrum
   * useless for future operations as it does an in-place conversion. Then
   * for each channel, compute interpolation samples using linear
   * interpolation on the amplitude and frequency
   */

   for bandNumber := 0; bandNumber < numberPartials; bandNumber++ {
     ampIndex := bandNumber * 2
     freqIndex := ampIndex + 1

     // Start where we left off, keep phase
     address := sineIndex[bandNumber]

     if polarSpectrum[ampIndex] == 0.0 {
       polarSpectrum[freqIndex] = float64(bandNumber) * cyclesBand
     } else {
       phaseDifference := polarSpectrum[freqIndex] - lastPhaseIn[bandNumber]
       lastPhaseIn[bandNumber] = polarSpectrum[freqIndex]

       // Unwrap phase differences
       for phaseDifference > pi {
         phaseDifference -= twoPi
       }

       for phaseDifference < -pi {
         phaseDifference += twoPi
       }

       // Convert to instantaneos frequency
       polarSpectrum[freqIndex] = phaseDifference * cyclesFrame + float64(bandNumber) * cyclesBand

       // Start with last amplitude
       amplitude := lastAmp[bandNumber]

       // Increment per sample to get to new amplitude
       ampIncrement := (polarSpectrum[ampIndex] - amplitude) * oneOvrInterp

       // Start with last frequency
       frequency := lastFreq[bandNumber]

       // Increment per sample to get to new frequency
       freqIncrement := (polarSpectrum[freqIndex] - frequency) * oneOvrInterp

       // Fill the output with one sine component
       for sample := 0; sample < interpolation; sample++ {
         // TODO: we are truncating a float to an int, should we round?
         output[sample] += amplitude * sineTable[int(address)]
         address += frequency

         // unwrap phase
         for address >= 8192 {
           address -= 8192
         }

         for address < 0 {
           address += 8192
         }

         amplitude += ampIncrement
         frequency += freqIncrement
       }
     }

     // save current values for next iteration
     lastFreq[bandNumber] = polarSpectrum[freqIndex]
     lastAmp[bandNumber] = polarSpectrum[ampIndex]
     sineIndex[bandNumber] = address
   }
 }
