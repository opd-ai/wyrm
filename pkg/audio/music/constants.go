// Package music provides an adaptive music system with genre-specific motifs.
// This file contains musical constants to reduce magic numbers in the adaptive music implementation.
package music

// Audio system constants.
const (
	// DefaultSampleRate is the audio sample rate in Hz (standard CD quality).
	DefaultSampleRate = 44100
	// DefaultCrossfadeDuration is the transition time between music states in seconds.
	// Per ROADMAP AC: Music transitions within 2s of entering combat.
	DefaultCrossfadeDuration = 2.0
	// CombatExitDelay is seconds after last enemy death before exiting combat music.
	// Per ROADMAP AC: Music reverts within 5s of last enemy death.
	CombatExitDelay = 5.0
	// MinVolume is the threshold below which a layer is considered inactive.
	MinVolume = 0.001
	// MaxVolume is the maximum volume level.
	MaxVolume = 1.0
	// CombatVolumeReduction is how much to reduce exploration volume during combat.
	CombatVolumeReduction = 0.3
)

// Musical frequency constants (Hz) - standard Western equal temperament.
const (
	// FreqA1 is the frequency of A1 (55 Hz).
	FreqA1 = 55.0
	// FreqE2 is the frequency of E2 (82.5 Hz).
	FreqE2 = 82.5
	// FreqA2 is the frequency of A2 (110 Hz).
	FreqA2 = 110.0
	// FreqE3 is the frequency of E3 (165 Hz).
	FreqE3 = 165.0
	// FreqA3 is the frequency of A3 (220 Hz).
	FreqA3 = 220.0
	// FreqE4 is the frequency of E4 (330 Hz).
	FreqE4 = 330.0
	// FreqA4 is the frequency of A4 (440 Hz) - standard concert pitch.
	FreqA4 = 440.0
)

// Interval constants - frequency ratios for musical intervals in equal temperament.
const (
	// Unison is the ratio for the same note.
	IntervalUnison = 1.0
	// MinorSecond is approximately a semitone up.
	IntervalMinorSecond = 1.059
	// MajorSecond is approximately a whole tone up.
	IntervalMajorSecond = 1.122
	// MinorThird is the ratio for a minor third interval.
	IntervalMinorThird = 1.189
	// MajorThird is the ratio for a major third interval.
	IntervalMajorThird = 1.25
	// PerfectFourth is the ratio for a perfect fourth interval.
	IntervalPerfectFourth = 1.333
	// Tritone is the ratio for an augmented fourth/diminished fifth.
	IntervalTritone = 1.414
	// PerfectFifth is the ratio for a perfect fifth interval.
	IntervalPerfectFifth = 1.5
	// MinorSixth is the ratio for a minor sixth interval.
	IntervalMinorSixth = 1.587
	// MajorSixth is the ratio for a major sixth interval.
	IntervalMajorSixth = 1.682
	// MinorSeventh is the ratio for a minor seventh interval.
	IntervalMinorSeventh = 1.782
	// MajorSeventh is the ratio for a major seventh interval.
	IntervalMajorSeventh = 1.888
	// Octave is the ratio for an octave interval.
	IntervalOctave = 2.0
	// IntervalDown5th is the ratio for a fifth below (3/4).
	IntervalDown5th = 0.75
	// IntervalDownMinor3rd is the ratio for a minor third below.
	IntervalDownMinor3rd = 0.833
	// IntervalDownMajor2nd is the ratio for a major second below.
	IntervalDownMajor2nd = 0.889
	// IntervalDownMinor2nd is the ratio for a minor second below.
	IntervalDownMinor2nd = 0.944
)

// Note duration constants in seconds.
const (
	// WholeNote at 60 BPM.
	WholeNote = 4.0
	// HalfNote at 60 BPM.
	HalfNote = 2.0
	// QuarterNote at 60 BPM.
	QuarterNote = 1.0
	// DottedQuarter is a quarter note plus an eighth.
	DottedQuarter = 1.5
	// EighthNote at 60 BPM.
	EighthNote = 0.5
	// SixteenthNote at 60 BPM.
	SixteenthNote = 0.25
	// ThirtySecondNote at 60 BPM.
	ThirtySecondNote = 0.125
)

// ADSR envelope constants (seconds) for audio synthesis.
const (
	// DefaultAttackTime is the envelope attack time.
	DefaultAttackTime = 0.02
	// DefaultDecayTime is the envelope decay time.
	DefaultDecayTime = 0.1
	// DefaultSustainLevel is the envelope sustain level (0.0 to 1.0).
	DefaultSustainLevel = 0.7
	// DefaultReleaseTime is the envelope release time.
	DefaultReleaseTime = 0.05
)

// Synthesis constants.
const (
	// DefaultWaveformAmplitude is the peak amplitude for generated waveforms.
	DefaultWaveformAmplitude = 0.3
	// TwoPi is 2*π for waveform calculations.
	TwoPi = 6.283185307179586
)
