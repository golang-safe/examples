// Command blinky blinks an LED on an MPS2 board, and says so on the serial port.
//
// It is the smallest complete SafeGo program that touches hardware, and it touches it the way
// a real program should: through the board's own package. There is not a register address in
// this file. `pkg.safego.dev/bsp/arm/mps2/*` owns those, reviewed once against the board's
// documentation, and every program on that board shares them.
//
// # Running it
//
//	safego build --target cortex-m3 -o out .
//	qemu-system-arm -machine mps2-an385 -cpu cortex-m3 -nographic -monitor none -kernel out/program.elf
//
// The LED is real; under QEMU nothing displays it, so the program also prints a character per
// toggle. **That output is the only observable effect on an emulator** — a blink you cannot
// see is indistinguishable from a program that hung.
package main

import (
	"pkg.safego.dev/bsp/arm/mps2/gpio"
	"pkg.safego.dev/bsp/arm/mps2/uart"
	"pkg.safego.dev/rt/time"
)

// ticker paces the blink.
//
// It is package-level because it must outlive the frame that made it — there is no heap — and
// it is a ticker rather than a delay loop because the scheduler puts the task to sleep between
// ticks against the board's hardware timebase. A busy-wait would burn the same wall-clock time
// and all of the power.
var ticker = time.NewTicker(200 * time.Millisecond)

// blinks is how many times to toggle before returning. A bounded program is easier to test
// than an endless one; a real one would loop forever, and the shape is otherwise identical.
const blinks int32 = 6

// blink toggles the LED on every tick and reports each toggle over the serial port.
func blink() {
	for i := int32(0); i < blinks; i++ {
		ticker.Wait()

		gpio.LED0.Toggle()

		var mark [1]byte
		if gpio.LED0.Read() {
			mark[0] = '*'
		} else {
			mark[0] = '.'
		}

		// The count is checked rather than discarded: a short write means the port stopped
		// accepting bytes, which is worth knowing even in an example.
		if n, err := uart.UART0.Write(mark[:]); err != nil || n != 1 {
			gpio.LED1.SetHigh()
		}
	}

	newline := [1]byte{'\n'}

	if _, err := uart.UART0.Write(newline[:]); err != nil {
		gpio.LED1.SetHigh()
	}
}

func main() {
	uart.Init()
	gpio.LED0.SetLow()

	//safego:task prio=1 stack=512
	go blink()
}
