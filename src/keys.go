package main

import (
	"time"
)

type Key int

const (
	NONE Key = -1

	UP    Key = 273
	DOWN  Key = 274
	LEFT  Key = 276
	RIGHT Key = 275

	SELECT Key = 305
	START  Key = 13

	X Key = 304
	Y Key = 308
	A Key = 32
	B Key = 306

	L  Key = 101
	L2 Key = 9

	R  Key = 116
	R2 Key = 8

	MENU Key = 27
)

func listenKeyPresses(channel chan Key) {
	for {
		value := Key(C_GetKeyPress())
		if value != NONE {
			channel <- value
		}

		time.Sleep(33 * time.Millisecond)
	}
}
