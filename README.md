# Overview

This is a Go port of the phase vocoding analysis/resynthesis routines from Tom Erbe's program ["SoundHack"](https://github.com/tomerbe/SoundHack).

Unlike the original SoundHack, this port does not have a UI and must be used on the command line.

# Differences from the original SoundHack

* gopvoc doesn't allow time stretching beyond the maximum or minimum as determined by the given inputs, see below. The way the "best" interpolation and decimation rates are determined is slightly different than the original SoundHack.
* gopvoc uses a slightly different vonn Hann window function than SoundHack
* gopvoc can process AIFF/WAV files with an arbitrary number of channels.
* gopvoc can only read and write AIFF and WAV files.
* gopvoc pitch shifting can only take a multiplier scale factor for pitch (octave lower is scale factor of 0.5, octave higher is 2.0, etc).
* gopvoc time stretching can only take a multiplier scale factor for time instead of a target output duration.
* gopvoc does not allow a scaling function, it only accepts a single value for scale factor.
* gopvoc handles output clipping differently than SoundHack. Before writing to disk, gopvoc clips any samples to the max or min allowed value for the given bit depth.
* gopvoc can analyze up to 8192 FFT bands.

# Commands

gopvoc has two modes of operation, time stretching and pitch shifting. They are invoked like so:

`./gopvoc time [options]`

`./gopvoc pitch [options]`

# Getting Help

To see options and defaults for each command:

`./gopvoc time -h`

`./gopvoc pitch -h`

# Flags and Options

Print gopvoc version:

`./gopvoc --version`

Both time stretching and pitch shifting use the following common set of flags:

Input AIFF/WAV file path (required):

`-i <path to input file>`

Output AIFF/WAV file path (required):

`-f <path to output file>`

Number of requested bands for FFT processing (must be one of: 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192):

`-b <number of bands>`

Overlap factor (must be one of 0.5, 1, 2, 4):

`-o <overlap>`

Scale factor (for time stretching, the amount to mutliply input duration by. For pitch shifting, the pitch shift multiplier):

`-s <scale factor>`

Windowing function for FFT processing (must be one of: rectangle, hamming, vonhann, kaiser, sinc, triangle, ramp):

`-w <window function name>`

Resynthesis gating minimum db level (optional):

`-ga <dBFS gate threshold>`

Resynthesis gating threshold db level below maximum (optional):

`-gt <dBFS gate threshold>`

Quiet flag (suppress stdout information and progress bar):

`-q`

Time stretching can do phase locking during resynthesis, to enable it:

`-p`

## Time Stretching

Time stretching is acheived via windowed FFT analysis of the input file, then resynthesis into the output file via [overlap add resynthesis](https://ccrma.stanford.edu/~jos/parshl/Overlap_Add_Synthesis.html).

Note that the original SoundHack could crash given certain program states based on extreme stretching multipliers. The maximum or minimum allowed `-s` scale multiplier is dependent on the FFT window size which in turn is dependent on the number of FFT bands requested by the `-b` flag and the overlap factor given by the `-o` flag. Unlike SoundHack, gopvoc will cap the multiplier within this limit instead of crashing due to a division by zero. Program output will indicate of your requested `-s <scale multiplier>` flag has been limited.

Example:

`./gopvoc time -i strings.aif -f strings_x10.aif -b 4096 -o 4 -s 10 -w kaiser`

The above example takes `strings.aif`, and stretches it to be 10 times the original length using 4096 FFT bands with an overlap factor of 4, using a kaiser windowing function

## Pitch Shifting

Pitch shifting is acheived via windowed FFT analysis of the input file, then resynthesis into the output file via [oscillator bank resynthesis](https://en.wikipedia.org/wiki/Additive_synthesis#Oscillator_bank_synthesis).

Example:

`./gopvoc pitch -i strings.aif -f strings_octavedown.aif -b 2048 -o 1 -s 0.5`

The above example takes `strings.aif`, and pitch shifts it down one octave (0.5 multipler of any given pitch in Hz is an octave lower) using 2048 FFT bands with an overlap factor of 1.

## Resynthesis Gating

After the forward FFT is taken, the data is passed through a spectral gate. The only operation available to the gate is to remove spectral data for a given FFT bin if the bin's amplitude does not meet one of the following thresholds. Either, both or none of these options may be passed:

**Resynthesis Gating Amplitude (db):**

`-ga -<db gate level>`

If an FFT bin's db amplitude does not meet the threshold given by the flag (expressed as db below 0dBFS), the bin's spectral data is removed.

Example:

`-ga -3`

Any spectral data with an amplitude less than -3db below 0dBFS will not be resynthesized.

**Resynthesis Gating Threshold (db) Below Maximum:**

`-gt -<db gate level>`

For each windowed FFT taken of the input, the maximum amplitude is found across all bins (frequencies). As above, a bin's spectral data is dropped, but only if the bin's amplitude does not meet the decibel threshold given, but compared against the maximum amplitude of the windowed FFT (as opposed to 0dBFS).

Example:

`-ga -10`

If in a given FFT analysis window, frequency bin #45 has the largest amplitude of all bins at -3dBFS, any frequency in the window with an amplitude below -13dbFS will be dropped. This is done for each FFT analysis window.

# Window Functions

Hamming window is the default window function. Because Hamming windows do not touch zero, some discontinuities are produced in the analysis and synthesis windowed data which may appear in some material as a "zippering" sound across channels. Try another window type like Kaiser, Sinc or von Hann which all touch zero.

# Build Instructions

* [Download and Install the Go language](https://go.dev/) for your system. Gopvoc has only been tested and built with Go 1.17.
* Clone this respository
* run `go build -o gopvoc`
* you should now have an executable `./gopvoc` in the current directory.

# Binaries

[Binaries have been cross-compiled](https://github.com/corporealfunk/gopvoc/releases) for OS X amd64 (Intel), arm64 (M1) as well as Linux (amd64) and Windows (amd64).
