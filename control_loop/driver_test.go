package main

import (
	"testing"

	"pkg.safego.dev/rt/volatile"
)

// A driver's decisions are the part most likely to be wrong and least likely to be
// exercised on hardware. These tests drive every branch of the UART driver with no board,
// no emulator and no cross compiler — which is what rt/volatile exists for.
//
// This file is never transpiled: it is a _test.go file, and the subset does not apply to
// one. It uses the modelled register bank directly, which a SafeGo program may not.

func TestWriteWaitsForRoom(t *testing.T) {
	volatile.Reset()

	// The transmitter reports full, so the driver must refuse rather than overwrite.
	volatile.Poke(uartBase+0x04, uint64(stateTxFull))

	if uartTryWrite('A') {
		t.Error("the driver wrote to a full transmitter")
	}

	if volatile.Peek(uartBase+0x00) != 0 {
		t.Error("the driver touched the data register when it should not have")
	}

	// Room appears.
	volatile.Poke(uartBase+0x04, 0)

	if !uartTryWrite('A') {
		t.Fatal("the driver refused an empty transmitter")
	}

	if got := volatile.Peek(uartBase + 0x00); got != 'A' {
		t.Errorf("the data register holds %#x, expected %#x", got, 'A')
	}
}

func TestReadReportsAnEmptyReceiver(t *testing.T) {
	volatile.Reset()

	if _, ok := uartTryRead(); ok {
		t.Error("the driver read from an empty receiver")
	}

	volatile.Poke(uartBase+0x04, uint64(stateRxFull))
	volatile.Poke(uartBase+0x00, 'Z')

	got, ok := uartTryRead()
	if !ok {
		t.Fatal("the driver missed a waiting byte")
	}

	if got != 'Z' {
		t.Errorf("read %#x, expected %#x", got, 'Z')
	}
}

// Enabling must set both directions and disturb nothing else: a driver that wrote the
// whole control register would silently clear whatever a previous stage configured.
func TestInitPreservesOtherBits(t *testing.T) {
	volatile.Reset()

	const foreign = 0x80

	volatile.Poke(uartBase+0x08, foreign)

	uartInit()

	got := volatile.Peek(uartBase + 0x08)

	if got&foreign == 0 {
		t.Error("init cleared a bit it does not own")
	}

	expected := uint64(controlTxEnable | controlRxEnable)
	if got&expected != expected {
		t.Errorf("init did not enable both directions: %#x", got)
	}
}
