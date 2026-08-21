// Command blinky-pico2 blinks the onboard LED on a Raspberry Pi Pico 2 (RP2350).
//
// It is the RP2350 sibling of ../blinky, which does the same thing on an MPS2 board under
// QEMU. There is not a register address in this file — pkg.safego.dev/bsp/raspberrypi/pico2/gpio
// owns those.
//
// # UNVERIFIED — this board port has never run
//
// Unlike ../blinky, there is no QEMU evidence here either: RP2350 is not in mainline QEMU's
// machine list. This example exists to be ready to flash the day a Pico 2 is in hand, not as
// a claim that it works.
//
// # Running it, once hardware bring-up is verified
//
//	safego build --target pico2 -o out .
//	<uf2 packaging tool, once one exists> out/program.elf out/program.uf2
//	<drag out/program.uf2 onto the Pico 2's mass-storage bootloader>
package main

import (
	"pkg.safego.dev/bsp/raspberrypi/pico2/gpio"
	"pkg.safego.dev/rt/time"
)

// ticker paces the blink — see ../blinky for why this is a ticker and not a delay loop.
var ticker = time.NewTicker(200 * time.Millisecond)

// blinks bounds the run, the same way ../blinky does, for the same reason: a bounded program
// is easier to reason about than an endless one, and the shape is otherwise identical.
const blinks int32 = 6

func blink() {
	for i := int32(0); i < blinks; i++ {
		ticker.Wait()

		gpio.LED0.Toggle()
	}
}

func main() {
	gpio.Init()
	gpio.LED0.SetLow()

	//safego:task prio=1 stack=512
	go blink()
}
