package main

import (
  "fmt"
  "os"
  "gopvoc/audioio"
  "gopvoc/pvoc"
  "gopvoc/cli"
  "github.com/schollz/progressbar/v3"
)

var Version = ""

func main() {
  // parse cli flags/arguments
  parsedArgs, err := cli.ParseFlags(os.Args, Version)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  // check if input file exists
  if _, err := os.Stat(parsedArgs.InputPath); err != nil {
    fmt.Fprintln(os.Stderr, "File does not exist:", parsedArgs.InputPath)
    os.Exit(1)
  }

  // setup the audioReader
  audioReader, err := audioio.NewAudioReader(parsedArgs.InputPath)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  // setup the Pvoc processor
  processor, err := pvoc.NewPvoc(
    parsedArgs.Bands,
    parsedArgs.Overlap,
    parsedArgs.Scale,
    parsedArgs.Operation,
    parsedArgs.PhaseLock,
    parsedArgs.WindowName,
    parsedArgs.GatingAmplitude,
    parsedArgs.GatingThreshold,
  )

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  if err = audioReader.Open(processor.Decimation); err != nil {
    fmt.Fprintln(os.Stderr, "Could not open AIFF file:", parsedArgs.InputPath)
    os.Exit(1)
  }

  defer audioReader.Close()

  if !parsedArgs.Quiet {
    fmt.Print(processor.String())

    fmt.Printf("%24s   %d\n", "Number of Channels:", audioReader.GetNumChans())
    fmt.Printf("%24s   %d\n", "Bit Depth:", audioReader.GetBitDepth())
    fmt.Printf("%24s   %d\n", "Sample Rate:", audioReader.GetSampleRate())
    fmt.Printf("%24s   %.2f\n", "Hz/FFT Band:", float64(audioReader.GetSampleRate()) / float64(processor.Bands) / 2.0)
    fmt.Printf("%24s   %.2f s\n", "Input Duration:", audioReader.GetDuration())

    if processor.Operation == pvoc.TimeStretch {
      fmt.Printf("%24s   %.2f s\n", "Output Duration:", audioReader.GetDuration() * processor.ScaleFactor)
    }
  }

  audioFile := audioio.AudioFile{
    Filepath: parsedArgs.OutputPath,
    NumChans: audioReader.GetNumChans(),
    SampleRate: audioReader.GetSampleRate(),
    BitDepth: audioReader.GetBitDepth(),
  }

  audioWriter, err := audioio.NewAudioWriter(audioFile)

  if err != nil {
    fmt.Fprintln(os.Stderr, "Could not create output audio file:", err)
    os.Exit(1)
  }

  if err = audioWriter.Create(processor.Interpolation); err != nil {
    fmt.Fprintln(os.Stderr, "Could not open audio file for writing:", parsedArgs.OutputPath)
    os.Exit(1)
  }

  defer audioWriter.Close()

  // progress will be a number 0-100
  progress := make(chan int)
  errors := make(chan error)
  done := make(chan bool)

  bar := progressbar.NewOptions(
    100,
    progressbar.OptionEnableColorCodes(true),
    progressbar.OptionSetDescription("processing..."),
    progressbar.OptionFullWidth(),
    progressbar.OptionSetTheme(progressbar.Theme{
      Saucer:        "[green]=[reset]",
      SaucerHead:    "[green]=[reset]",
      SaucerPadding: " ",
      BarStart:      "[",
      BarEnd:        "]",
    }),
  )

  go processor.Run(
    audioReader,
    audioWriter,
    progress,
    errors,
    done,
  )

  // wait for messages
  wait := true
  for wait {
    select {
    case err := <- errors:
      fmt.Fprintln(os.Stderr, "\n >>> Processing error:", err, " <<<\n")
      os.Exit(1)
    case curProgress := <-progress:
      if !parsedArgs.Quiet {
        bar.Set(curProgress)
      }
    case <- done:
      if !parsedArgs.Quiet {
        fmt.Println("\n\nDone!")
      }
      wait = false
    }
  }
}
