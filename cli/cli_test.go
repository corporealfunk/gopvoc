package cli

import (
	"fmt"
	"gopvoc/pvoc"
	. "gopvoc/testing_utilities"
	"path/filepath"
	"testing"
)

func TestParseOutputFilePath(t *testing.T) {
  fixturesPath, _ := filepath.Abs("../fixtures")

  tests := map[string]struct{
    outputPath    string
    parsedArgs    *Arguments
    expected      string
    hasError      bool
  }{
    "full file path exists": {
      outputPath: "../fixtures/tmp/myfilename.aif",
      expected: fmt.Sprintf("%s/%s", fixturesPath, "tmp/myfilename.aif"),
      parsedArgs: &Arguments{},
      hasError: false,
    },
    "full file path does not exists": {
      outputPath: "../fixtures/tmp/nothere.aif",
      expected: fmt.Sprintf("%s/%s", fixturesPath, "tmp/nothere.aif"),
      parsedArgs: &Arguments{},
      hasError: false,
    },
    "full file path but directory does not exist": {
      outputPath: "../fixtures/tmpz/myfilename.aif",
      expected: fmt.Sprintf("%s/%s", fixturesPath, "tmp/myfilename.aif"),
      parsedArgs: &Arguments{},
      hasError: true,
    },
    "directory only, base path does not exist": {
      outputPath: "../fixtures/tmpz/another",
      expected: fmt.Sprintf("%s/%s", fixturesPath, "invalid"),
      parsedArgs: &Arguments{
        Bands: 4096,
        Overlap: 4,
        Scale: 15,
        Operation: pvoc.TimeStretch,
        InputPath: "../fixtures/sine_1_chan.aif",
        OutputPath: "../fixtures/tmp",
        PhaseLock: false,
        WindowName: "hamming",
      },
      hasError: true,
    },
    "directory only, base path exists, time most defaults": {
      outputPath: "../fixtures/tmp",
      expected: fmt.Sprintf("%s/%s", fixturesPath, "tmp/out-ts100.aif"),
      parsedArgs: &Arguments{
        Bands: 4096,
        Overlap: 1,
        Scale: 100,
        Operation: pvoc.TimeStretch,
        InputPath: "../fixtures/out.aif",
        OutputPath: "../fixtures/",
        PhaseLock: false,
        WindowName: "hamming",
      },
      hasError: false,
    },
    "directory only, base path exists, time no defaults": {
      outputPath: "../fixtures/tmp",
      expected: fmt.Sprintf("%s/%s", fixturesPath, "tmp/out-ts0125-o05-b8-kaiser-ga13-gt20-p.aif"),
      parsedArgs: &Arguments{
        Bands: 8,
        Overlap: 0.5,
        Scale: 0.125,
        Operation: pvoc.TimeStretch,
        InputPath: "../fixtures/out.aif",
        OutputPath: "../fixtures/",
        PhaseLock: true,
        WindowName: "kaiser",
        GatingAmplitude: -13,
        GatingThreshold: -20,
      },
      hasError: false,
    },
  }

  for name, test := range tests {
    t.Run(name, func(t *testing.T){
      output, err := parseOutputFilePath(
        test.outputPath,
        test.parsedArgs,
      )

      if !test.hasError {
        Equals(
          t,
          test.expected,
          output,
        )
      } else {
        Assert(
          t,
          err != nil,
          "err should not be nil",
        )
      }
    })
  }
}
