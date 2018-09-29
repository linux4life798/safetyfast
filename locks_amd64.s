// +build 386 amd64
// HLE instructions can work on 386 or amd64

#include "textflag.h"

TEXT ·Pause(SB),NOPTR|NOSPLIT,$0-0
    PAUSE
    RET

// Atomically write 1 to val while returning the old value.
// func Lock1XCHG8(val *int8) (old int8)
TEXT ·Lock1XCHG8(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    MOVB $1, AX
    LOCK
    XCHGB AX, (CX)
    MOVB AX, old+8(FP)
    RET

// Atomically write 1 to val while returning the old value.
// func Lock1XCHG32(val *int32) (old int32)
TEXT ·Lock1XCHG32(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    MOVL $1, AX
    LOCK
    XCHGL AX, (CX)
    MOVL AX, old+8(FP)
    RET

// Atomically write 1 to val while returning the old value.
// func Lock1XCHG64(val *int64) (old int64)
TEXT ·Lock1XCHG64(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    MOVL $1, AX
    LOCK
    XCHGQ AX, (CX)
    MOVQ AX, old+8(FP)
    RET

// SpinLock implements a basic spin lock that waits forever until the lock
// can be acquired.
// It assumes the lock has been acquired when it successfully writes a 1 to val,
// while the original value was 0.
// This implementation spins on a read only and then uses the XCHG instruction
// in order to claim the lock. The loop makes use of the PAUSE hint instruction.
// func SpinLock(val *int32)
TEXT ·SpinLock(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
tryread:
    MOVL (CX), BX
    TESTL BX, BX
    JE tryacquire
    PAUSE
    JMP tryread
tryacquire:
    MOVL $1, AX
    LOCK
    XCHGL AX, (CX)
    TESTL AX, AX
    JNE tryread
    RET

// SpinCountLock implements a basic spin lock that tries to acquire the lock
// only attempts times.
// It assumes the lock has been acquired when it successfully writes a 1 to val,
// while the original value was 0.
// This implementation spins on a read only and then uses the XCHG instruction
// in order to claim the lock. The loop makes use of the PAUSE hint instruction.
// func SpinCountLock(val, attempts *int32)
TEXT ·SpinCountLock(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    // Load attempt counter in DX
    MOVQ attempts+8(FP), R8
    MOVL (R8), DX
tryread:
    MOVL (CX), BX
    TESTL BX, BX
    JE tryacquire
    PAUSE
    DECL DX
    // If DX != 0, abort
    JNE tryread
    JMP abort
tryacquire:
    MOVL $1, AX
    LOCK
    XCHGL AX, (CX)
    TESTL AX, AX
    JNE tryread
abort:
    // Write back attempt counter
    MOVL DX, (R8)
    RET

// HLETryLock attempts only once to acquire the lock by writing a 1 to val
// using HLE primitives. This function returns a 0 if the lock was acquired.
// func HLETryLock(val *int32) (old int32)
TEXT ·HLETryLock(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    MOVL $1, AX
    XACQUIRE
    XCHGL AX, (CX)
    MOVL AX, old+8(FP)
    RET

// HLEUnlock writes a 0 to val to indicate the lock has been released
// using HLE primitives
// func HLEUnlock(val *int32)
TEXT ·HLEUnlock(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    MOVL $0, AX
    XRELEASE
    MOVL AX, (CX)
    RET

// func HLESpinLock(val *int32)
TEXT ·HLESpinLock(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
tryread:
    MOVL (CX), BX
    TESTL BX, BX
    JE tryacquire
    PAUSE
    JMP tryread
tryacquire:
    MOVL $1, AX
    XACQUIRE
    XCHGL AX, (CX)
    TESTL AX, AX
    JNE tryread
    RET

// Note: Argument attempts must be greater than 0
// func HLESpinCountLock(val, attempts *int32)
TEXT ·HLESpinCountLock(SB),NOPTR|NOSPLIT,$0
    MOVQ val+0(FP), CX
    // Load attempt counter in DX
    MOVQ attempts+8(FP), R8
    MOVL (R8), DX
tryread:
    MOVL (CX), BX
    TESTL BX, BX
    JE tryacquire
    PAUSE
    DECL DX
    // If DX != 0, abort
    JNE tryread
    JMP abort
tryacquire:
    MOVL $1, AX
    XACQUIRE
    XCHGL AX, (CX)
    TESTL AX, AX
    JNE tryread
abort:
    // Write back attempt counter
    MOVL DX, (R8)
    RET
