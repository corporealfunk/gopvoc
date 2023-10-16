package cli

import(
  "flag"
  "fmt"
  "os"
  "path/filepath"
  "gopvoc/pvoc"
  "strings"
  "math"
)

type Arguments struct {
  Bands int
  Overlap float64
  Scale float64
  Operation int
  Quiet bool
  InputPath string
  OutputPath string
  PhaseLock bool
  WindowName string
  GatingAmplitude float64
  GatingThreshold float64
}

func parseOutputFilePath(outputFile string, parsedArgs *Arguments) (string, error) {
  fullPath, _ := filepath.Abs(outputFile)
  pathDir := filepath.Dir(fullPath)

  // does the base directory open
  _, err := os.Open(pathDir)
  if err != nil {
    return "", err
  }

  // does the full path open
  file, err := os.Open(fullPath)

  // it doesn't open, just return it
  if err != nil {
    return fullPath, nil
  }

  fileInfo, err := file.Stat()

  // opens but doesn't stat, return
  if err != nil {
    return "", err
  }

  // it opens, if it is not a directory return it
  if !fileInfo.IsDir() {
    return fullPath, nil
  }

  // it is a directory that exists, create a filename
  fileName := filepath.Base(parsedArgs.InputPath)
  ext := filepath.Ext(fileName)
  operation := "t"

  if parsedArgs.Operation == pvoc.PitchShift {
    operation = "p"
  }

  overlap := ""
  if parsedArgs.Overlap != 1.0 {
    overlap = fmt.Sprintf("-o%g", parsedArgs.Overlap)
  }

  bands := ""

  if parsedArgs.Bands != 4096 {
    bands = fmt.Sprintf("-b%d", parsedArgs.Bands)
  }

  window := ""

  if parsedArgs.WindowName != "hamming" {
    window = fmt.Sprintf("-%s", parsedArgs.WindowName)
  }

  gatingA := ""

  if parsedArgs.GatingAmplitude != 0 {
    gatingA = fmt.Sprintf("-ga%g", math.Abs(parsedArgs.GatingAmplitude))
  }

  gatingT := ""

  if parsedArgs.GatingThreshold != 0 {
    gatingT = fmt.Sprintf("-gt%g", math.Abs(parsedArgs.GatingThreshold))
  }

  phaseLock := ""
  if parsedArgs.PhaseLock {
    phaseLock = "-p"
  }

  builtName := strings.Replace(
    fmt.Sprintf(
      "%s-%ss%g%s%s%s%s%s%s",
      strings.TrimSuffix(fileName, ext),
      operation,
      parsedArgs.Scale,
      overlap,
      bands,
      window,
      gatingA,
      gatingT,
      phaseLock,
    ),
    ".",
    "",
    -1,
  )

  builtName = fmt.Sprintf("%s%s", builtName, ext)

  return filepath.Join(fullPath, builtName), nil
}

func ParseFlags(args []string, version string) (*Arguments, error) {
  var flgVersion bool
  flag.BoolVar(&flgVersion, "version", false, "print version and exit")

  flag.Parse()

  if flgVersion {
    fmt.Printf("version: %s\n", version)
    os.Exit(0)
  }

  cmdError := fmt.Errorf("usage: gopvoc [--version] <command> <args>\n\nAvailable Commands:\n\n    time    time stretch input AIFF/WAV file\n    pitch   pitch shift input AIFF/WAV file\n\nFor specific command options:\n\ngopvoc <command> -h\n\n")

  if len(args) < 2 {
    return nil, cmdError
  }

  // time stretch flags
  timeCmd := flag.NewFlagSet("time", flag.ExitOnError)
  timeInput := timeCmd.String("i", "", "input file: path to input AIFF/WAV")
  timeScale := timeCmd.Float64("s", 1.0, "scale factor: time scale multiplier")
  timeBands := timeCmd.Int("b", 4096, "bands: number of FFT bands to use during processing. Must be a power of two between 2 to 8192 inclusive")
  timeOverlap := timeCmd.Float64("o", 1.0, "overlap: overlap factor, allowed values: 0.5, 1, 2, 4")
  timePhaseLock := timeCmd.Bool("p", false, "phase lock flag: enable phase locking during resynthesis")
  timeWindowName := timeCmd.String("w", "hamming", "window: windowing function to use, one of: " + pvoc.WindowNamesString())
  timeGatingAmplitude := timeCmd.Float64("ga", 0.0, "resynthesis gating amplitude (db): amplitude below 0db under which an FFT frequency is removed from the spectrum.")
  timeGatingThreshold := timeCmd.Float64("gt", 0.0, "resynthesis gating threshold (db) below maximum: any FFT frequency bin with an amplitude this far below the maximum amplitude of all bins in that FFT window will get removed.")
  timeQuiet := timeCmd.Bool("q", false, "quiet flag: suppress informational output")
  timeOutput := timeCmd.String("f", "", "output file or directory: Provide a path to an AIFF/WAV file. If only a directory is specified, the output file will be automatically named. In both cases, file will be overwritten if it exists.")

  // pitch flags
  pitchCmd := flag.NewFlagSet("pitch", flag.ExitOnError)
  pitchInput := pitchCmd.String("i", "", "input file: path to input AIFF/WAV")
  pitchScale := pitchCmd.Float64("s", 1.0, "scale factor: time scale multiplier")
  pitchBands := pitchCmd.Int("b", 4096, "bands: number of FFT bands to use during processing. Must be a power of two between 2 to 8192 inclusive")
  pitchOverlap := pitchCmd.Float64("o", 1.0, "overlap: overlap factor, allowed values: 0.5, 1, 2, 4")
  pitchWindowName := pitchCmd.String("w", "hamming", "window: windowing function to use, one of: " + pvoc.WindowNamesString())
  pitchGatingAmplitude := pitchCmd.Float64("ga", 0.0, "resynthesis gating amplitude (db): amplitude below 0db under which an FFT frequency is removed from the spectrum.")
  pitchGatingThreshold := pitchCmd.Float64("gt", 0.0, "resynthesis gating threshold (db) below maximum: any FFT frequency bin with an amplitude this far below the maximum amplitude of all bins in that FFT window will get removed.")
  pitchQuiet := pitchCmd.Bool("q", false, "quiet flag: suppress informational output")
  pitchOutput := pitchCmd.String("f", "", "output file or directory: Provide a path to an AIFF/WAV file. If only a directory is specified, the output file will be automatically named. In both cases, file will be overwritten if it exists.")

  parsedArgs := &Arguments{ }

  switch args[1] {
  case "time":
    timeCmd.Parse(os.Args[2:])

    if len(*timeInput) == 0 {
      return nil, fmt.Errorf("Required argument missing:\n\n-i <path to input file> is required, for help:\n\ngopvoc time -h\n\n")
    }

    parsedArgs.Operation = pvoc.TimeStretch
    parsedArgs.InputPath, _ = filepath.Abs(*timeInput)
    parsedArgs.Scale = *timeScale
    parsedArgs.Bands = *timeBands
    parsedArgs.Overlap = *timeOverlap
    parsedArgs.PhaseLock = *timePhaseLock
    parsedArgs.WindowName = *timeWindowName
    parsedArgs.GatingAmplitude = *timeGatingAmplitude
    parsedArgs.GatingThreshold = *timeGatingThreshold
    parsedArgs.Quiet = *timeQuiet

    if len(*timeOutput) == 0 {
      return nil, fmt.Errorf("Required argument missing:\n\n-f <path to output file or directory> is required, for help:\n\ngopvoc time -h\n\n")
    }

    parsedFilePath, err := parseOutputFilePath(*timeOutput, parsedArgs)
    if err != nil {
      return nil, err
    }
    parsedArgs.OutputPath = parsedFilePath
  case "pitch":
    pitchCmd.Parse(os.Args[2:])
    parsedArgs.Operation = pvoc.PitchShift

    if len(*pitchInput) == 0 {
      return nil, fmt.Errorf("Required argument missing:\n\n-i <path to input file> is required, for help:\n\ngopvoc time -h\n\n")
    }

    parsedArgs.InputPath, _ = filepath.Abs(*pitchInput)
    parsedArgs.Scale = *pitchScale
    parsedArgs.Bands = *pitchBands
    parsedArgs.Overlap = *pitchOverlap
    parsedArgs.WindowName = *pitchWindowName
    parsedArgs.GatingAmplitude = *pitchGatingAmplitude
    parsedArgs.GatingThreshold = *pitchGatingThreshold
    parsedArgs.Quiet = *pitchQuiet

    if len(*pitchOutput) == 0 {
      return nil, fmt.Errorf("Required argument missing:\n\n-f <path to output file> is required, for help:\n\ngopvoc time -h\n\n")
    }

    parsedFilePath, err := parseOutputFilePath(*pitchOutput, parsedArgs)
    if err != nil {
      return nil, err
    }
    parsedArgs.OutputPath = parsedFilePath
  default:
    return nil, cmdError
  }

  return parsedArgs, nil
}
