package main

import (
  "fmt"
  "os"
  "gopvoc/audioio"
  "gopvoc/pvoc"
  "gopvoc/cli"
  "github.com/schollz/progressbar/v3"
)

func main() {
  // parse cli flags/arguments
  parsedArgs, err := cli.ParseFlags(os.Args)

  if err != nil {
    fmt.Fprintln(os.Stderr, err)
    os.Exit(1)
  }

  // check if input file exists
  if _, err := os.Stat(parsedArgs.InputPath); err != nil {
    fmt.Fprintln(os.Stderr, "File does not exist:", parsedArgs.InputPath)
    os.Exit(1)
  }

  // setup the aiffReader
  aiffReader := &audioio.AiffReader{Filepath: parsedArgs.InputPath}

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

  if err = aiffReader.Open(processor.Decimation); err != nil {
    fmt.Fprintln(os.Stderr, "Could not open AIFF file:", parsedArgs.InputPath)
    os.Exit(1)
  }

  defer aiffReader.Close()

  if !parsedArgs.Quiet {
    fmt.Print(processor.String())

    fmt.Printf("%24s   %d\n", "Number of Channels:", aiffReader.NumChans)
    fmt.Printf("%24s   %d\n", "Bit Depth:", aiffReader.BitDepth)
    fmt.Printf("%24s   %f s\n", "Input Duration:", aiffReader.Duration)

    if processor.Operation == pvoc.TimeStretch {
      fmt.Printf("%24s   %f s\n", "Output Duration:", aiffReader.Duration * processor.ScaleFactor)
    }
  }

  aiffWriter := &audioio.AiffWriter{
    Filepath: parsedArgs.OutputPath,
    NumChans: aiffReader.NumChans,
    SampleRate: aiffReader.SampleRate,
    BitDepth: aiffReader.BitDepth,
  }

  if err = aiffWriter.Create(processor.Interpolation); err != nil {
    fmt.Fprintln(os.Stderr, "Could not open AIFF file for writing:", parsedArgs.OutputPath)
    os.Exit(1)
  }

  defer aiffWriter.Close()

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
    aiffReader,
    aiffWriter,
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
