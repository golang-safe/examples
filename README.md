# SafeGo examples

Worked SafeGo programs. Each is a module of its own, built with the SafeGo compiler and run on
QEMU — no hardware required.

| Example | What it shows |
| :--- | :--- |
| [`blinky`](blinky) | The smallest program that touches hardware: a memory-mapped LED, a hardware-paced ticker and one task |
| [`control_loop`](control_loop) | A fixed-rate control task, a mutex-protected setpoint and an ISR |

## Running one

```sh
cd blinky
safego build --target cortex-m3 -o out .
qemu-system-arm -machine mps2-an385 -cpu cortex-m3 -nographic -monitor none -kernel out/program.elf
```

`blinky` prints `*` and `.` as it toggles. **That output is the point on an emulator**: QEMU
models the MPS2's LED register but shows nothing for it, and a blink you cannot observe is
indistinguishable from a program that hung.

## Only QEMU, for now

Both examples target the MPS2-AN385, which QEMU models and which the compiler's own tests run
against. The board ports for physical hardware are not finished — see the SafeGo repository —
so an example that claimed to run on a Pico or an STM32 would be claiming more than the
toolchain can currently deliver.

## Licence

Business Source License 1.1 — see `LICENSE`.
