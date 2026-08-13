// Command blinky drives a real LED on a real board, on a real clock.
//
// It is the smallest complete SafeGo program that touches hardware: a package-level register
// for the LED, a package-level ticker for the period, and one task that toggles the first
// every time the second fires. Everything a larger program does is a variation on this.
//
// # The board
//
// An ARM MPS2-AN385, which QEMU models, so this runs without hardware:
//
//	qemu-system-arm -machine mps2-an385 -cpu cortex-m3 -nographic -kernel program.elf
//
// The LED is bit 0 of the CMSDK FPGA I/O LED register. On a bench board it is a light; under
// QEMU nothing displays it, which is why the task also writes a character to UART0 on every
// toggle. **That character is the only observable effect under emulation**, and it is what the
// test asserts — a blink you cannot see is indistinguishable from a program that hung.
//
// # Why it is shaped this way
//
// The register and the ticker are package-level because they must be: a register address has
// to be a compile-time constant so the build can place it and a reviewer can check it against
// a datasheet (C-001), and a ticker has to outlive the frame that made it because there is no
// heap (C-002). The subset refuses both otherwise, which is the compiler insisting on a shape
// that was always correct for a program with no dynamic memory.
package main

import (
	"pkg.safego.dev/rt/time"
	"pkg.safego.dev/rt/volatile"
)

// CMSDK FPGA I/O, which carries the user LEDs on the MPS2 boards. The LED bank is a
// peripheral of the FPGA image rather than of the processor: a Cortex-M3 on another device has
// none of it, which is why the address lives in the program and not in the runtime.
const fpgaioBase uintptr = 0x40028000

// CMSDK UART0. Bit 0 of the state register is set while the transmit holding register is
// still full, so a byte is written only once it clears.
const uartBase uintptr = 0x40004000

const (
	// stateTxFull is the transmit-busy bit of the UART's state register.
	stateTxFull uint32 = 0x01

	// ledBit is the first user LED.
	ledBit uint32 = 0x01

	// txWaitLimit bounds the wait for the transmit register. A condition-only loop has no
	// bound the analyzer can compute (R-401), and an unbounded wait on a peripheral that
	// never clears is a hang rather than an error — so the wait gives up instead.
	txWaitLimit int32 = 100000

	// blinks is how many times to toggle before returning. A bounded program is easier to
	// test than an endless one, and the shape is the same either way.
	blinks int32 = 6

	// baudDiv is the smallest divisor the CMSDK UART accepts. Without it the peripheral
	// never reports room in the transmit register and nothing is ever sent.
	baudDiv uint32 = 16

	// ctrlTxEnable turns the transmitter on.
	ctrlTxEnable uint32 = 0x01
)

var (
	// led is the FPGA I/O LED bank.
	led = volatile.Reg32(fpgaioBase + 0x00)

	// uartData and uartState are the UART's data and status registers.
	uartData  = volatile.Reg32(uartBase + 0x00)
	uartState = volatile.Reg32(uartBase + 0x04)

	// uartCtrl enables the transmitter and uartBaudDiv sets the divisor. Nothing does this
	// for us: the board's startup code brings up the clock and the timers, not the UART.
	uartCtrl    = volatile.Reg32(uartBase + 0x08)
	uartBaudDiv = volatile.Reg32(uartBase + 0x10)

	// ticker paces the blink. It is driven by the board's hardware timebase through the
	// scheduler, so the task is asleep between blinks rather than spinning — which is what
	// makes this a program a battery could run.
	ticker = time.NewTicker(200 * time.Millisecond)

	// done carries the blink count out of the task, so main can observe that it finished.
	done = make(chan uint8, 1)
)

// putc writes one byte to the UART, waiting for room with a bounded loop.
func putc(c uint8) {
	for w := int32(0); w < txWaitLimit; w++ {
		if !uartState.HasBits(stateTxFull) {
			break
		}
	}

	uartData.Store(uint32(c))
}

// blink toggles the LED on every tick and reports each toggle over the UART.
//
// The loop is bounded, so the task returns; a task that ran forever would be equally valid and
// is the more usual shape, but this one is meant to be run by a test that waits for it to end.
func blink() {
	var on bool

	for i := int32(0); i < blinks; i++ {
		ticker.Wait()

		on = !on

		if on {
			led.SetBits(ledBit)
			putc('*')
		} else {
			led.ClearBits(ledBit)
			putc('.')
		}
	}

	putc('\n')

	done <- 1
}

func main() {
	uartBaudDiv.Store(baudDiv)
	uartCtrl.Store(ctrlTxEnable)

	led.ClearBits(ledBit)

	//safego:task prio=1 stack=512
	go blink()
}
