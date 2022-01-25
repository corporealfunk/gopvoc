package cli

import(
  "flag"
  "fmt"
  "os"
  "path/filepath"
  "gopvoc/pvoc"
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

func ParseFlags(args []string) (*Arguments, error) {
  cmdError := fmt.Errorf("usage: gopvoc <command> <args>\n\nAvailable Commands:\n\n    time    time stretch input AIFF file\n    pitch   pitch shift input AIFF file\n\nFor specific command options:\n\ngopvoc <command> -h\n\n")

  if len(args) < 2 {
    return nil, cmdError
  }

  // time stretch flags
  timeCmd := flag.NewFlagSet("time", flag.ExitOnError)
  timeInput := timeCmd.String("i", "", "input file: path to input AIFF")
  timeScale := timeCmd.Float64("s", 1.0, "scale factor: time scale multiplier")
  timeBands := timeCmd.Int("b", 4096, "bands: number of FFT bands to use during processing. Must be a power of two between 2 to 4096 inclusive")
  timeOverlap := timeCmd.Float64("o", 1.0, "overlap: overlap factor, allowed values: 0.5, 1, 2, 4")
  timePhaseLock := timeCmd.Bool("p", false, "phase lock flag: enable phase locking during resynthesis")
  timeWindowName := timeCmd.String("w", "hamming", "window: windowing function to use, one of: " + pvoc.WindowNamesString())
  timeGatingAmplitude := timeCmd.Float64("ga", 0.0, "resynthesis gating amplitude (db): amplitude below 0db under which an FFT frequency is removed from the spectrum.")
  timeGatingThreshold := timeCmd.Float64("gt", 0.0, "resynthesis gating threshold (db) below maximum: any FFT frequency bin with an amplitude this far below the maximum amplitude of all bins in that FFT window will get removed.")
  timeQuiet := timeCmd.Bool("q", false, "quiet flag: suppress informational output")
  timeOutput := timeCmd.String("f", "", "output file: path to write output AIFF. It will be overwritten if it exists")

  // pitch flags
  pitchCmd := flag.NewFlagSet("pitch", flag.ExitOnError)
  pitchInput := pitchCmd.String("i", "", "input file: path to input AIFF")
  pitchScale := pitchCmd.Float64("s", 1.0, "scale factor: time scale multiplier")
  pitchBands := pitchCmd.Int("b", 4096, "bands: number of FFT bands to use during processing. Must be a power of two between 2 to 4096 inclusive")
  pitchOverlap := pitchCmd.Float64("o", 1.0, "overlap: overlap factor, allowed values: 0.5, 1, 2, 4")
  pitchWindowName := pitchCmd.String("w", "hamming", "window: windowing function to use, one of: " + pvoc.WindowNamesString())
  pitchGatingAmplitude := pitchCmd.Float64("ga", 0.0, "resynthesis gating amplitude (db): amplitude below 0db under which an FFT frequency is removed from the spectrum.")
  pitchGatingThreshold := pitchCmd.Float64("gt", 0.0, "resynthesis gating threshold (db) below maximum: any FFT frequency bin with an amplitude this far below the maximum amplitude of all bins in that FFT window will get removed.")
  pitchQuiet := pitchCmd.Bool("q", false, "quiet flag: suppress informational output")
  pitchOutput := pitchCmd.String("f", "", "output file: path to write output AIFF. It will be overwritten if it exists")

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
      return nil, fmt.Errorf("Required argument missing:\n\n-f <path to output file> is required, for help:\n\ngopvoc time -h\n\n")
    }

    parsedArgs.OutputPath, _ = filepath.Abs(*timeOutput)
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

    parsedArgs.OutputPath, _ = filepath.Abs(*pitchOutput)
  default:
    return nil, cmdError
  }

  return parsedArgs, nil
}
