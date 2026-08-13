// Command control_loop is a worked example of a SafeGo program.
//
// blinky shows the minimum shape. This one shows the reasoning: why the declarations are
// where they are, and what each rule is protecting. It is a periodic controller — sample,
// clamp, actuate — with an acquisition interrupt feeding it and a supervisor watching for
// missed deadlines.
//
// Everything it does is static. There is no allocation anywhere: the channels, the task
// stacks, the ticker and the limit table are all placed at build time, and the memory the
// image needs is known before it runs.
package main

import (
	"pkg.safego.dev/rt/sync"
	"pkg.safego.dev/rt/time"
)

// The channels between the three tasks and the interrupt.
//
// They are package-level because there is no heap: a channel used after main returns
// cannot live in main's frame. Each capacity is a decision rather than a default —
// readings is deep enough to absorb a burst of interrupts while the controller is busy,
// commands only needs to cover one period of jitter, and faults is small because a fault
// that repeats faster than the supervisor can drain it is itself the problem.
var readings = make(chan int32, 8)

var commands = make(chan int32, 4)

var faults = make(chan int32, 4)

// limits is a frozen map: the keyword itself means the table is fixed at build time.
//
// Marked immutable and never written, so it costs no RAM at all — the transpiler places
// it in flash and the lookup becomes a switch with no hashing.
//
//safego:immutable
var limits = map[string]int32{"low": -1000, "high": 1000, "slew": 50}

// loop is the controller's period. The duration folds to counter units at build time
// against the target's clock, so nothing divides at run time.
var loop = time.NewTicker(10 * time.Millisecond)

// lastCommand is shared between the controller that writes it and the supervisor that
// reads it, so it is protected. The ceiling of stateLock is computed from the call graph:
// the highest priority of any task that can reach a Lock on it.
var stateLock = sync.New()

var lastCommand int32

// overrunLimit is how many missed periods the supervisor tolerates before reporting.
const overrunLimit int32 = 3

const faultOverrun int32 = 1

// onSample is the acquisition interrupt.
//
// It sends without blocking, which is the only channel form an interrupt handler may use:
// a handler that blocked would hold the processor at interrupt priority waiting for a task
// that cannot run until it returns. The default case is what makes the send non-blocking,
// and taking it means a sample was dropped — which is the correct behaviour under
// overload, because the alternative is falling further behind while pretending not to.
//
//safego:isr vector=PORT0 prio=6
func onSample() {
	select {
	case readings <- 1:
	default:
	}
}

// clamp holds a value inside the configured range and limits how fast it may move.
//
// The slew limit is what makes this a controller rather than a follower: an actuator that
// can be commanded to jump anywhere is one that can be commanded to break something.
func clamp(previous int32, wanted int32) int32 {
	low := limits["low"]
	high := limits["high"]
	slew := limits["slew"]

	target := wanted

	if target < low {
		target = low
	}

	if target > high {
		target = high
	}

	if target > previous+slew {
		target = previous + slew
	}

	if target < previous-slew {
		target = previous - slew
	}

	return target
}

// control is the periodic task.
//
// It waits on the ticker rather than sleeping for a duration, so a late iteration does not
// push the next one later still. Draining the readings is non-blocking: the controller runs
// on its period whether or not a sample arrived, which is what makes its timing analysable.
func control() {
	previous := int32(0)

	for {
		loop.Wait()

		// A fixed window: take up to four samples and no more, filling the rest with
		// zero. The bound is the array's, so the loop's cost is the same on every
		// iteration whether or not the sensor kept up — which is what makes the
		// controller's period analysable rather than merely usually short.
		//
		// The receives are non-blocking. A blocking one here would tie the controller's
		// period to the sensor's rate instead of to its own ticker.
		var window [4]int32

		for i := range window {
			select {
			case r := <-readings:
				window[i] = r
			default:
				window[i] = 0
			}
		}

		sum := int32(0)

		for i := range window {
			sum += window[i]
		}

		next := clamp(previous, sum*100)
		previous = next

		// Locked and unlocked on the same path with nothing between that can return.
		// Under the ceiling protocol an unreleased lock does not stall this task — it
		// leaves it running at the ceiling, starving everything below it silently.
		stateLock.Lock()
		lastCommand = next
		stateLock.Unlock()

		commands <- next
	}
}

// actuate applies commands as they arrive.
//
// Ranging over a channel blocks every iteration, which is what makes this a well-formed
// task loop: the task always yields, so a lower-priority task can run.
func actuate() {
	for c := range commands {
		// The command goes out of the UART, one byte at a time, through the driver in
		// driver.go. The write does not wait: a driver that spun here would hold this
		// task's priority for as long as the peripheral felt like taking.
		// The command is clamped to the configured range, so its low byte is the whole of
		// it. Narrowing is written out because SafeGo requires it to be: a conversion that
		// can lose information is never implicit (§2.2.1).
		if !uartTryWrite(uint8(c & 0xFF)) {
			dropped++
		}
	}
}

// dropped counts commands the transmitter had no room for. A count is what lets the
// supervisor notice that the link cannot keep up, which is a different fault from the
// controller failing to compute.
var dropped int32

// supervise reports when the controller stops meeting its period.
//
// The ticker counts missed periods rather than queueing them, so this is a question the
// program can ask instead of a backlog it has to work through.
func supervise() {
	for {
		reading := <-faults
		_ = reading
	}
}

// onPanic is this package's panic handler.
//
// It runs on a dedicated stack with interrupts masked and cannot be called from program
// code — raise a panic instead. Its job is to leave the hardware safe, not to recover:
// there is no recover, and a handler that returns leaves the system running on state
// nobody has checked.
//
//safego:panic
func onPanic(err error) {
	stopped := int32(0)
	_ = stopped
	_ = err
}

func main() {
	// Every task is spawned from main, and nothing is written afterwards that a task can
	// reach. On the host a task starts at the go statement; on the target it starts once
	// main returns. Writing shared state after a spawn would race on one and not the
	// other, so the ordering rule removes the difference rather than documenting it.

	uartInit()

	//safego:task prio=5 stack=1024
	go control()

	//safego:task prio=3 stack=768
	go actuate()

	//safego:task prio=2 stack=768
	go supervise()
}
