module safego.example/blinky-pico2

go 1.25.0

require (
	pkg.safego.dev/bsp/raspberrypi v0.0.0
	pkg.safego.dev/rt v0.0.0-20260810093946-29f8792115d6
)

// pkg.safego.dev/bsp/raspberrypi is not published yet; the board package is developed
// alongside, the same way ../blinky replaces pkg.safego.dev/bsp/arm.
replace pkg.safego.dev/bsp/raspberrypi => ../../bsp-raspberrypi
