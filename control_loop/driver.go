package main

import "pkg.safego.dev/rt/volatile"

// The MPS2 board's UART, as a driver written the way any peripheral is.
//
// The registers are names for addresses. On the target each access below becomes a
// volatile read or write of exactly the declared width; on the host the same code runs
// against a modelled register bank, which is what makes the logic testable without a board
// — see driver_test.go, which drives every branch here with no hardware present.
//
// That is the whole argument for D1 in one file: the bit patterns and the state machine are
// the part most likely to be wrong and least likely to be exercised on a bench.
const uartBase uintptr = 0x40004000

var uartData = volatile.Reg32(uartBase + 0x00)

var uartState = volatile.Reg32(uartBase + 0x04)

var uartControl = volatile.Reg32(uartBase + 0x08)

// The bits this driver uses, named rather than spelled at each site: a bare 0x02 in the
// middle of a condition is a value nobody can check against a datasheet.
const (
	stateTxFull uint32 = 0x01
	stateRxFull uint32 = 0x02

	controlTxEnable uint32 = 0x01
	controlRxEnable uint32 = 0x02
)

// uartInit enables both directions.
func uartInit() {
	uartControl.SetBits(controlTxEnable | controlRxEnable)
}

// uartTryWrite sends one byte if the transmitter has room, and reports whether it did.
//
// It does not wait. A driver that spun here would hold whatever priority it was called at
// for as long as the peripheral felt like taking, which is a timing property no analysis
// could bound — so the decision to retry belongs to the caller.
func uartTryWrite(b uint8) bool {
	if uartState.HasBits(stateTxFull) {
		return false
	}

	uartData.Store(uint32(b))

	return true
}

// uartTryRead takes one byte if the receiver has one.
func uartTryRead() (uint8, bool) {
	if !uartState.HasBits(stateRxFull) {
		return 0, false
	}

	// The data register holds one byte in its low bits; the rest are reserved.
	return uint8(uartData.Load() & 0xFF), true
}
