# SafeGo examples

Worked SafeGo programs. Each is a module of its own, built with the SafeGo compiler. Most run
on QEMU, no hardware required — `blinky-pico2` is the one exception; see below.

| Example | What it shows |
| :--- | :--- |
| [`blinky`](blinky) | The smallest program that touches hardware: a memory-mapped LED, a hardware-paced ticker and one task |
| [`control_loop`](control_loop) | A fixed-rate control task, a mutex-protected setpoint and an ISR |
| [`blinky-pico2`](blinky-pico2) | The same idea as `blinky`, targeting a real board (Raspberry Pi Pico 2) instead of QEMU — **builds, has never run; see below** |

## Running one

```sh
cd blinky
safego build --target cortex-m3 -o out .
qemu-system-arm -machine mps2-an385 -cpu cortex-m3 -nographic -monitor none -kernel out/program.elf
```

`blinky` prints `*` and `.` as it toggles. **That output is the point on an emulator**: QEMU
models the MPS2's LED register but shows nothing for it, and a blink you cannot observe is
indistinguishable from a program that hung.

## Only QEMU has ever actually run one

`blinky` and `control_loop` target the MPS2-AN385, which QEMU models and which the compiler's
own tests run against. Every other board port, `pico2` included, is written against its
datasheet and SVD but has not been through a compiler on real silicon, let alone flashed —
see the SafeGo repository's `ports/` for what "unverified" means for each one.

`blinky-pico2` is the exception to "only QEMU": it is not runnable at all right now, on
anything. RP2350 is not in mainline QEMU's machine list, so there is no emulator evidence
either — only that the build succeeds and the register offsets have been checked by hand
against the SVD.

```sh
cd blinky-pico2
safego build --target pico2 -o out .
arm-none-eabi-objcopy -O binary out/program.elf out/program.bin
bin2uf2 -family rp2350-arm-ns -base 0x10000000 out/program.bin out/program.uf2
# drag out/program.uf2 onto the Pico 2's mass-storage bootloader, once someone has one in hand
```

`bin2uf2` is built from the SafeGo repository's `cmd/bin2uf2`.

## Licence

Business Source License 1.1 — see `LICENSE`.
