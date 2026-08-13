module safego.example/blinky

go 1.25.0

require (
	pkg.safego.dev/bsp/arm v0.0.0
	pkg.safego.dev/rt v0.0.0-20260810093946-29f8792115d6
)

require pkg.safego.dev/hal v0.0.0-20260813121838-0c59a23c8aaa // indirect

// pkg.safego.dev/bsp is not published yet; the board package is developed alongside.
replace pkg.safego.dev/bsp/arm => ../../bsp-arm
